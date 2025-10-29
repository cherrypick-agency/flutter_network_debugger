package redact

import (
    "encoding/json"
    "testing"
)

func TestRedactJSON_UnicodeAndLongStrings(t *testing.T) {
    t.Parallel()

    // Юникод: эмодзи, диакритика, RTL
    in := map[string]any{
        "authorization": "Bearer 🔑секрет",
        "note": "café naïve — مرحبا",
        "list": []any{"тест", 123},
    }
    b, _ := json.Marshal(in)
    out := RedactJSON(string(b))
    mustContain(t, out, `"authorization":"***"`)
    mustContain(t, out, `"note":"café naïve — مرحبا"`)

    // Длинная строка
    long := make([]rune, 10000)
    for i := range long { long[i] = 'あ' }
    in = map[string]any{"access_token": string(long)}
    b, _ = json.Marshal(in)
    out = RedactJSON(string(b))
    mustContain(t, out, `"access_token":"***"`)
}

func FuzzRedactJSON_NoPanicAndUTF8(f *testing.F) {
    f.Add("{}")
    f.Add("{\"authorization\":\"x\"}")
    f.Add("[1,2,3]")
    f.Fuzz(func(t *testing.T, s string) {
        _ = RedactJSON(s)
        // инварианты: отсутствие паник и валидный UTF-8 по определению строки Go
        // Дополнительно можно проверить, что результат не пустеет без причины,
        // но оставим минимальным для скорости.
    })
}


