import 'package:flutter_test/flutter_test.dart';
import 'package:hex_viewer/hex_viewer.dart';

void main() {
  group('HexFormatter', () {
    group('formatByte', () {
      test('formats zero correctly', () {
        expect(HexFormatter.formatByte(0), '00');
      });

      test('formats max byte correctly', () {
        expect(HexFormatter.formatByte(255), 'FF');
      });

      test('formats single digit hex', () {
        expect(HexFormatter.formatByte(10), '0A');
      });

      test('formats double digit hex', () {
        expect(HexFormatter.formatByte(16), '10');
      });
    });

    group('formatOffset', () {
      test('formats zero offset', () {
        expect(HexFormatter.formatOffset(0), '00000000');
      });

      test('formats non-zero offset', () {
        expect(HexFormatter.formatOffset(256), '00000100');
      });

      test('formats large offset', () {
        expect(HexFormatter.formatOffset(0xABCDEF), '00ABCDEF');
      });

      test('respects custom width', () {
        expect(HexFormatter.formatOffset(10, width: 4), '000A');
      });
    });

    group('formatAscii', () {
      test('formats printable ASCII', () {
        expect(HexFormatter.formatAscii(65), 'A'); // 'A'
        expect(HexFormatter.formatAscii(97), 'a'); // 'a'
        expect(HexFormatter.formatAscii(48), '0'); // '0'
      });

      test('formats space', () {
        expect(HexFormatter.formatAscii(32), ' '); // space
      });

      test('formats null byte as dot', () {
        expect(HexFormatter.formatAscii(0), '·');
      });

      test('formats control characters as dot', () {
        expect(HexFormatter.formatAscii(1), '·');
        expect(HexFormatter.formatAscii(10), '·'); // newline
        expect(HexFormatter.formatAscii(13), '·'); // carriage return
      });

      test('formats extended ASCII as dot', () {
        expect(HexFormatter.formatAscii(128), '·');
        expect(HexFormatter.formatAscii(255), '·');
      });
    });

    group('formatLine', () {
      test('formats empty data', () {
        final config = const HexConfig();
        final line = HexFormatter.formatLine([], 0, config);

        expect(line.offset, 0);
        expect(line.bytes, isEmpty);
        expect(line.hexString, isEmpty);
        expect(line.asciiString, isEmpty);
      });

      test('formats single byte', () {
        final config = const HexConfig();
        final data = [0x41]; // 'A'
        final line = HexFormatter.formatLine(data, 0, config);

        expect(line.offset, 0);
        expect(line.bytes, [0x41]);
        expect(line.hexString, '41');
        expect(line.asciiString, 'A');
      });

      test('formats full line without grouping', () {
        final config = const HexConfig(groupSize: 0);
        final data = List.generate(16, (i) => i);
        final line = HexFormatter.formatLine(data, 0, config);

        expect(line.offset, 0);
        expect(line.bytes.length, 16);
        expect(
          line.hexString,
          '00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F',
        );
      });

      test('formats full line with grouping (8-8)', () {
        final config = const HexConfig(groupSize: 8);
        final data = List.generate(16, (i) => i);
        final line = HexFormatter.formatLine(data, 0, config);

        expect(line.offset, 0);
        expect(line.bytes.length, 16);
        expect(line.hexString.contains('|'), isTrue);
        final parts = line.hexString.split(' | ');
        expect(parts.length, 2);
      });

      test('formats partial line', () {
        final config = const HexConfig(bytesPerLine: 16);
        final data = [0x48, 0x65, 0x6C, 0x6C, 0x6F]; // "Hello"
        final line = HexFormatter.formatLine(data, 0, config);

        expect(line.offset, 0);
        expect(line.bytes.length, 5);
        expect(line.asciiString, 'Hello');
      });

      test('handles offset beyond data', () {
        final config = const HexConfig();
        final data = [0x41];
        final line = HexFormatter.formatLine(data, 100, config);

        expect(line.bytes, isEmpty);
      });
    });

    group('calculateLineCount', () {
      test('calculates zero for empty data', () {
        expect(HexFormatter.calculateLineCount(0, 16), 0);
      });

      test('calculates one line for partial data', () {
        expect(HexFormatter.calculateLineCount(5, 16), 1);
      });

      test('calculates exact lines', () {
        expect(HexFormatter.calculateLineCount(32, 16), 2);
      });

      test('calculates with remainder', () {
        expect(HexFormatter.calculateLineCount(33, 16), 3);
      });
    });

    group('formatSelectionAsHex', () {
      test('formats empty selection', () {
        expect(HexFormatter.formatSelectionAsHex([]), isEmpty);
      });

      test('formats single byte', () {
        expect(HexFormatter.formatSelectionAsHex([0xAB]), 'AB');
      });

      test('formats multiple bytes', () {
        expect(
          HexFormatter.formatSelectionAsHex([0x48, 0x65, 0x6C]),
          '48 65 6C',
        );
      });
    });

    group('formatSelectionAsText', () {
      test('formats printable text', () {
        expect(
          HexFormatter.formatSelectionAsText([0x48, 0x65, 0x6C, 0x6C, 0x6F]),
          'Hello',
        );
      });

      test('formats mixed content', () {
        expect(HexFormatter.formatSelectionAsText([0x48, 0x00, 0x65]), 'H·e');
      });
    });
  });
}
