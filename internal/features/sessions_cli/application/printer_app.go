package application

import (
	"context"
	"io"
	core "network-debugger/internal/domain"
	"network-debugger/internal/features/sessions_cli/domain"
	"network-debugger/internal/features/sessions_cli/ports"
	"strconv"
	"strings"
)

// PrinterApp — coordinator: event subscription, data loading, printing.
type PrinterApp struct {
	Events ports.EventStream
	Sess   ports.SessionQuery
	Tx     ports.HTTPTxQuery
	Frames ports.FramesQuery
	Rend   interface {
		RenderHTTP(io.Writer, domain.HTTPView, domain.Options) error
	}
}

func (p *PrinterApp) Run(ctx context.Context, w io.Writer, opts domain.Options) error {
	opts.EnsureFieldsFromPreset()
	ch, cancel, err := p.Events.Subscribe(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-ch:
			if !ok {
				return nil
			}
			switch ev.Type {
			case ports.EventHTTPTxAdded:
				// Try to get the transaction directly by ID (in-memory fast path)
				if finder, ok := p.Tx.(interface {
					FindHTTPTransaction(context.Context, string, string) (core.HTTPTransaction, bool)
				}); ok {
					tx2, ok2 := finder.FindHTTPTransaction(ctx, ev.SessionID, ev.Ref)
					if ok2 {
						vm := domain.HTTPView{
							SessionID: ev.SessionID,
							StartedAt: tx2.StartedAt,
							Method:    tx2.Method,
							URL:       tx2.URL,
							Status:    tx2.Status,
							ReqSize:   tx2.ReqSize,
							RespSize:  tx2.RespSize,
							TotalMs:   tx2.Timings.Total,
						}
						_ = p.Rend.RenderHTTP(w, vm, opts)
						continue
					}
				}
				// Fallback: take a small "tail" and search for the needed ID
				txs, _, err := p.Tx.ListHTTPTransactions(ctx, ev.SessionID, "", 256)
				if err != nil || len(txs) == 0 {
					continue
				}
				var tx core.HTTPTransaction
				found := false
				for i := len(txs) - 1; i >= 0; i-- {
					if txs[i].ID == ev.Ref {
						tx = txs[i]
						found = true
						break
					}
				}
				if !found {
					// if not found in the tail — don't make noise, wait for the next event
					continue
				}
				// Filtering (simple substring by URL/method/status)
				if f := strings.TrimSpace(opts.Filter); f != "" {
					if !(strings.Contains(strings.ToLower(tx.URL), strings.ToLower(f)) || strings.Contains(strings.ToLower(tx.Method), strings.ToLower(f)) || strings.Contains(strings.ToLower(strconv.Itoa(tx.Status)), strings.ToLower(f))) {
						continue
					}
				}
				vm := domain.HTTPView{
					SessionID: ev.SessionID,
					StartedAt: tx.StartedAt,
					Method:    tx.Method,
					URL:       tx.URL,
					Status:    tx.Status,
					ReqSize:   tx.ReqSize,
					RespSize:  tx.RespSize,
					TotalMs:   tx.Timings.Total,
				}
				_ = p.Rend.RenderHTTP(w, vm, opts)
			default:
				// ignore other events for MVP
			}
		}
	}
}
