package httpapi

import (
	"bufio"
	"bytes"
	"compress/flate"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"network-debugger/internal/domain"
	"network-debugger/pkg/shared/id"
)

// pipeWSRaw читает по одному WebSocket кадру из src и переправляет его в dst,
// формируя превью для первых previewMaxBytes байт полезной нагрузки.
// Минимальная поддержка RFC6455: не разбираем фрагментацию/расширения.
func (d *Deps) pipeWSRaw(sessionID string, src net.Conn, dst net.Conn, direction domain.Direction) {
	defer func() {
		_ = src.Close()
		_ = dst.Close()
		// Завершение сессии, если другая сторона уже не закрыла
		ctx := contextWithNoCancel()
		if sess, ok, _ := d.Svc.Get(ctx, sessionID); !ok || sess.ClosedAt == nil {
			_ = d.Svc.SetClosed(ctx, sessionID, time.Now().UTC(), nil)
			d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
			d.Metrics.ActiveSessions.Dec()
		}
	}()

	br := bufio.NewReader(src)
	for {
		op, size, preview, err := d.forwardOneWSFrame(br, dst)
		if err != nil {
			if err != io.EOF {
				_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
			}
			return
		}
		// Логируем кадр
		fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: direction, Opcode: op, Size: size, Preview: preview}
		_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
		d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
		d.Metrics.FramesTotal.WithLabelValues(string(direction), string(op)).Inc()
	}
}

// pipeWSMessages собирает WS-сообщения из последовательности кадров (учитывая continuation/FIN),
// пишет исходные кадры в dst и логирует превью на уровне сообщения. Поддерживает best-effort
// распаковку permessage-deflate для превью, если включено в конфиге.
func (d *Deps) pipeWSMessages(sessionID string, src net.Conn, dst net.Conn, direction domain.Direction) {
	defer func() {
		_ = src.Close()
		_ = dst.Close()
		ctx := contextWithNoCancel()
		if sess, ok, _ := d.Svc.Get(ctx, sessionID); !ok || sess.ClosedAt == nil {
			_ = d.Svc.SetClosed(ctx, sessionID, time.Now().UTC(), nil)
			d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
			d.Metrics.ActiveSessions.Dec()
		}
	}()

	br := bufio.NewReader(src)
	// Оборачиваем writer, чтобы равномерно замедлять запись
	pacedDst := &wsPacedWriter{d: d, w: dst, dir: direction}
	for {
		opMain, msgSize, preview, rawText, bodyFile, err := d.forwardOneWSMessage(br, pacedDst)
		if err != nil {
			if err != io.EOF {
				_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
			}
			return
		}
		fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: direction, Opcode: opMain, Size: msgSize, Preview: preview}
		if bodyFile != "" {
			fr.BodyFile = bodyFile
		}
		_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
		d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
		d.Metrics.FramesTotal.WithLabelValues(string(direction), string(opMain)).Inc()

		// SIO parse for text messages (forward mode) — единый путь
		if opMain == domain.OpcodeText {
			if len(rawText) > 0 {
				_ = d.recordSIOIfAny(sessionID, rawText, fr.ID)
			} else if len(preview) > 0 {
				_ = d.recordSIOIfAny(sessionID, preview, fr.ID)
			}
		}
	}
}

// forwardOneWSFrame читает один WS‑кадр из r и пишет его как есть в w, возвращая
// (opcode, payloadSize, preview). Поддерживает MASK и расширенные длины.
func (d *Deps) forwardOneWSFrame(r *bufio.Reader, w io.Writer) (domain.Opcode, int, string, error) {
	// Заголовок (минимум 2 байта)
	b0, err := r.ReadByte()
	if err != nil {
		return domain.OpcodeBinary, 0, "", err
	}
	b1, err := r.ReadByte()
	if err != nil {
		return domain.OpcodeBinary, 0, "", err
	}
	// Опкод
	opCode := b0 & 0x0F
	op := opcodeFromWS(opCode)

	// MASK и длина
	mask := (b1 & 0x80) != 0
	plen7 := int(b1 & 0x7F)
	payloadLen := 0
	var extLen [8]byte
	switch plen7 {
	case 126:
		if _, err := io.ReadFull(r, extLen[:2]); err != nil {
			return op, 0, "", err
		}
		payloadLen = int(binary.BigEndian.Uint16(extLen[:2]))
	case 127:
		if _, err := io.ReadFull(r, extLen[:]); err != nil {
			return op, 0, "", err
		}
		// Ограничим до int (практически всегда укладывается)
		ln := binary.BigEndian.Uint64(extLen[:])
		if ln > 1<<31-1 {
			ln = 1<<31 - 1
		}
		payloadLen = int(ln)
	default:
		payloadLen = plen7
	}

	// Пишем заголовок в dst
	if _, err := w.Write([]byte{b0, b1}); err != nil {
		return op, 0, "", err
	}
	if plen7 == 126 {
		if _, err := w.Write(extLen[:2]); err != nil {
			return op, 0, "", err
		}
	} else if plen7 == 127 {
		if _, err := w.Write(extLen[:]); err != nil {
			return op, 0, "", err
		}
	}

	// Masking key (если есть)
	var maskKey [4]byte
	if mask {
		if _, err := io.ReadFull(r, maskKey[:]); err != nil {
			return op, 0, "", err
		}
		if _, err := w.Write(maskKey[:]); err != nil {
			return op, 0, "", err
		}
	}

	// Копируем payload чанками; формируем превью по первым previewMaxBytes
	// Для превью размаскируем если нужно
	previewMax := d.Cfg.PreviewMaxBytes
	if previewMax <= 0 {
		previewMax = 4096
	}
	previewBuf := make([]byte, 0, min(previewMax, payloadLen))
	buf := make([]byte, 32*1024)
	remaining := payloadLen
	offset := 0
	for remaining > 0 {
		toRead := len(buf)
		if toRead > remaining {
			toRead = remaining
		}
		n, err := io.ReadFull(r, buf[:toRead])
		if err != nil {
			return op, 0, "", err
		}
		// Пишем как есть в dst
		if _, err := w.Write(buf[:n]); err != nil {
			return op, 0, "", err
		}
		// Формируем превью
		// Маскирование применяется только к клиентским кадрам, но мы смотрим на бит MASK
		take := n
		if len(previewBuf) < previewMax {
			need := previewMax - len(previewBuf)
			if take > need {
				take = need
			}
			chunk := buf[:take]
			if mask {
				// демаскируем копию
				tmp := make([]byte, take)
				for i := 0; i < take; i++ {
					tmp[i] = chunk[i] ^ maskKey[(offset+i)&3]
				}
				previewBuf = append(previewBuf, tmp...)
			} else {
				previewBuf = append(previewBuf, chunk...)
			}
		}
		remaining -= n
		offset += n
	}

	// Сборка превью
	var preview string
	if op == domain.OpcodeText {
		preview = string(previewBuf)
		if previewMax > 0 && len(preview) > previewMax {
			preview = preview[:previewMax]
		}
	} else {
		preview = hexPreview(previewBuf)
	}

	return op, payloadLen, preview, nil
}

// forwardOneWSMessage читает один логический WS-месседж (Text/Binary/Control),
// потоково пробрасывая кадры в dst. Возвращает opcode сообщения, полный размер payload
// и превью (для text — строка, для binary — hex), а также "rawText" — исходный текст
// для дальнейшего парсинга SIO (может быть усечён). Для control кадров возврат аналогичен кадру.
func (d *Deps) forwardOneWSMessage(r *bufio.Reader, w io.Writer) (domain.Opcode, int, string, string, string, error) {
	// Читаем первый кадр сообщения
	// Скопируем первый байт для проверки FIN/RSV1/opcode
	b0, err := r.ReadByte()
	if err != nil {
		return domain.OpcodeBinary, 0, "", "", "", err
	}
	b1, err := r.ReadByte()
	if err != nil {
		return domain.OpcodeBinary, 0, "", "", "", err
	}
	fin := (b0 & 0x80) != 0
	rsv1 := (b0 & 0x40) != 0
	opCode := b0 & 0x0F
	opMain := opcodeFromWS(opCode)

	mask := (b1 & 0x80) != 0
	plen7 := int(b1 & 0x7F)
	payloadLen := 0
	var extLen [8]byte
	switch plen7 {
	case 126:
		if _, err := io.ReadFull(r, extLen[:2]); err != nil {
			return opMain, 0, "", "", "", err
		}
		payloadLen = int(binary.BigEndian.Uint16(extLen[:2]))
	case 127:
		if _, err := io.ReadFull(r, extLen[:]); err != nil {
			return opMain, 0, "", "", "", err
		}
		ln := binary.BigEndian.Uint64(extLen[:])
		if ln > 1<<31-1 {
			ln = 1<<31 - 1
		}
		payloadLen = int(ln)
	default:
		payloadLen = plen7
	}

	// Пишем заголовок первого кадра
	if _, err := w.Write([]byte{b0, b1}); err != nil {
		return opMain, 0, "", "", "", err
	}
	if plen7 == 126 {
		if _, err := w.Write(extLen[:2]); err != nil {
			return opMain, 0, "", "", "", err
		}
	} else if plen7 == 127 {
		if _, err := w.Write(extLen[:]); err != nil {
			return opMain, 0, "", "", "", err
		}
	}
	var maskKey [4]byte
	if mask {
		if _, err := io.ReadFull(r, maskKey[:]); err != nil {
			return opMain, 0, "", "", "", err
		}
		if _, err := w.Write(maskKey[:]); err != nil {
			return opMain, 0, "", "", "", err
		}
	}

	wsPreviewMax := d.Cfg.WSPreviewMaxBytes
	if wsPreviewMax <= 0 {
		wsPreviewMax = d.Cfg.PreviewMaxBytes
		if wsPreviewMax <= 0 {
			wsPreviewMax = 4096
		}
	}
	// Собираем сообщение: для превью держим отдельный буфер (демаскированный),
	// а для permessage-deflate — сберём сжатые байты (первые WS_BODY_MAX_BYTES)
	previewBuf := make([]byte, 0, min(wsPreviewMax, payloadLen))
	// подготовим фаловую шпулю для полного дампа, если включено
	var spool *os.File
	var spoolPath string
	var spoolRemain int64
	if d.Cfg.WSCaptureBodies && payloadLen > 0 && (opMain == domain.OpcodeText || opMain == domain.OpcodeBinary) {
		spool, spoolPath = d.openWSFile()
		if spool != nil {
			spoolRemain = int64(d.Cfg.WSBodyMaxBytes)
			if spoolRemain <= 0 {
				spoolRemain = int64(payloadLen)
			}
		}
	}
	var compressedBuf []byte
	if rsv1 && d.Cfg.WSDeflatePreview {
		capBytes := d.Cfg.WSBodyMaxBytes
		if capBytes <= 0 {
			capBytes = 1 << 20 // 1MB по умолчанию
		}
		if payloadLen < capBytes {
			capBytes = payloadLen
		}
		compressedBuf = make([]byte, 0, capBytes)
	}

	// helper для приёма тела кадра
	readPayload := func(rem int, applyMask bool, key [4]byte, offsetBase int) (int, error) {
		buf := make([]byte, 32*1024)
		off := offsetBase
		for rem > 0 {
			toRead := len(buf)
			if toRead > rem {
				toRead = rem
			}
			n, err := io.ReadFull(r, buf[:toRead])
			if err != nil {
				return off, err
			}
			// forward оригинальные байты
			if _, err := w.Write(buf[:n]); err != nil {
				return off, err
			}
			// превью
			if len(previewBuf) < wsPreviewMax {
				take := n
				need := wsPreviewMax - len(previewBuf)
				if take > need {
					take = need
				}
				chunk := buf[:take]
				if applyMask {
					tmp := make([]byte, take)
					for i := 0; i < take; i++ {
						tmp[i] = chunk[i] ^ key[(off+i)&3]
					}
					previewBuf = append(previewBuf, tmp...)
				} else {
					previewBuf = append(previewBuf, chunk...)
				}
			}
			if rsv1 && d.Cfg.WSDeflatePreview && len(compressedBuf) < cap(compressedBuf) {
				take := n
				if take > cap(compressedBuf)-len(compressedBuf) {
					take = cap(compressedBuf) - len(compressedBuf)
				}
				if take > 0 {
					compressedBuf = append(compressedBuf, buf[:take]...)
				}
			}
			// дамп в файл (демаскированный payload на уровне сообщения)
			if spool != nil && spoolRemain > 0 {
				toWrite := n
				if int64(toWrite) > spoolRemain {
					toWrite = int(spoolRemain)
				}
				if toWrite > 0 {
					// для маски — демаскируем копию
					if mask {
						tmp := make([]byte, toWrite)
						for i := 0; i < toWrite; i++ {
							tmp[i] = buf[i] ^ maskKey[(off+i)&3]
						}
						_, _ = spool.Write(tmp)
					} else {
						_, _ = spool.Write(buf[:toWrite])
					}
					spoolRemain -= int64(toWrite)
				}
			}
			rem -= n
			off += n
		}
		return off, nil
	}

	// читаем payload первого кадра
	off, err := readPayload(payloadLen, mask, maskKey, 0)
	if err != nil {
		return opMain, 0, "", "", "", err
	}
	totalSize := payloadLen

	// Если это control‑кадр или fin=true — сообщение закончено
	if fin || opCode == 0x8 || opCode == 0x9 || opCode == 0xA {
		preview, raw := d.buildWSPreview(opMain, rsv1, previewBuf, compressedBuf)
		if spool != nil {
			_ = spool.Close()
		}
		return opMain, totalSize, preview, raw, spoolPath, nil
	}

	// Иначе — собираем continuation кадры
	for {
		// следующий заголовок
		b0c, err := r.ReadByte()
		if err != nil {
			return opMain, totalSize, "", "", "", err
		}
		b1c, err := r.ReadByte()
		if err != nil {
			return opMain, totalSize, "", "", "", err
		}
		fin = (b0c & 0x80) != 0
		// permessage-deflate: RSV1 ставится только в первом кадре сообщения
		// опкод должен быть 0x0 (continuation)
		plen7 = int(b1c & 0x7F)
		mask = (b1c & 0x80) != 0
		payloadLen = 0
		switch plen7 {
		case 126:
			if _, err := io.ReadFull(r, extLen[:2]); err != nil {
				return opMain, totalSize, "", "", "", err
			}
			payloadLen = int(binary.BigEndian.Uint16(extLen[:2]))
		case 127:
			if _, err := io.ReadFull(r, extLen[:]); err != nil {
				return opMain, totalSize, "", "", "", err
			}
			ln := binary.BigEndian.Uint64(extLen[:])
			if ln > 1<<31-1 {
				ln = 1<<31 - 1
			}
			payloadLen = int(ln)
		default:
			payloadLen = plen7
		}
		// forward header
		if _, err := w.Write([]byte{b0c, b1c}); err != nil {
			return opMain, totalSize, "", "", "", err
		}
		if plen7 == 126 {
			if _, err := w.Write(extLen[:2]); err != nil {
				return opMain, totalSize, "", "", "", err
			}
		} else if plen7 == 127 {
			if _, err := w.Write(extLen[:]); err != nil {
				return opMain, totalSize, "", "", "", err
			}
		}
		if mask {
			if _, err := io.ReadFull(r, maskKey[:]); err != nil {
				return opMain, totalSize, "", "", "", err
			}
			if _, err := w.Write(maskKey[:]); err != nil {
				return opMain, totalSize, "", "", "", err
			}
		}
		off, err = readPayload(payloadLen, mask, maskKey, off)
		if err != nil {
			return opMain, totalSize, "", "", "", err
		}
		totalSize += payloadLen
		if fin {
			preview, raw := d.buildWSPreview(opMain, rsv1, previewBuf, compressedBuf)
			if spool != nil {
				_ = spool.Close()
			}
			return opMain, totalSize, preview, raw, spoolPath, nil
		}
	}
}

func (d *Deps) buildWSPreview(op domain.Opcode, rsv1 bool, previewBuf []byte, compressedBuf []byte) (string, string) {
	// Если текст и есть сжатая копия — попробуем распаковать для превью/сырого текста
	if op == domain.OpcodeText && rsv1 && d.Cfg.WSDeflatePreview && len(compressedBuf) > 0 {
		// permessage-deflate требует добавления zlib tail 0x00 0x00 0xFF 0xFF
		data := append([]byte(nil), compressedBuf...)
		data = append(data, 0x00, 0x00, 0xFF, 0xFF)
		zr := flate.NewReader(bytes.NewReader(data))
		if zr != nil {
			defer zr.Close()
			// ограничим чтение
			lim := d.Cfg.WSPreviewMaxBytes
			if lim <= 0 {
				lim = d.Cfg.PreviewMaxBytes
				if lim <= 0 {
					lim = 4096
				}
			}
			out := make([]byte, 0, lim)
			buf := make([]byte, 1024)
			for len(out) < lim {
				n, err := zr.Read(buf)
				if n > 0 {
					need := lim - len(out)
					if n > need {
						n = need
					}
					out = append(out, buf[:n]...)
				}
				if err != nil {
					break
				}
			}
			// Возвращаем превью/сырой как строку
			return string(out), string(out)
		}
	}
	// Без распаковки
	if op == domain.OpcodeText {
		s := string(previewBuf)
		return s, s
	}
	return hexPreview(previewBuf), ""
}

// openWSFile создаёт временный файл в BodySpoolDir и возвращает файл и абсолютный путь
func (d *Deps) openWSFile() (*os.File, string) {
	dir := d.Cfg.BodySpoolDir
	if dir == "" {
		dir = os.TempDir()
	}
	_ = os.MkdirAll(dir, 0o755)
	f, err := os.CreateTemp(dir, "gpx-ws-*.bin")
	if err != nil {
		return nil, ""
	}
	abs, _ := filepath.Abs(f.Name())
	return f, abs
}

func opcodeFromWS(code byte) domain.Opcode {
	switch code {
	case 0x1:
		return domain.OpcodeText
	case 0x2:
		return domain.OpcodeBinary
	case 0x8:
		return domain.OpcodeClose
	case 0x9:
		return domain.OpcodePing
	case 0xA:
		return domain.OpcodePong
	default:
		return domain.OpcodeBinary
	}
}

func hexPreview(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	// Ограничим до 256 байт как в wsproxy
	if len(b) > 256 {
		b = b[:256]
	}
	dst := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(dst, b)
	return string(dst)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// wsPacedWriter добавляет задержку записи пропорционально размеру, чтобы симулировать
// ограничение пропускной способности для WS при ручной прокачке кадров.
type wsPacedWriter struct {
	d   *Deps
	w   io.Writer
	dir domain.Direction
}

func (p *wsPacedWriter) Write(b []byte) (int, error) {
	// Пауза до записи, чтобы не перегонять лишнего за тик
	if p != nil && p.d != nil && p.d.Cfg.ThrottleEnabled {
		p.d.throttleSleepWS(p.dir, len(b))
	}
	return p.w.Write(b)
}
