package domain

type Opcode string

const (
    OpcodeText   Opcode = "text"
    OpcodeBinary Opcode = "binary"
    OpcodePing   Opcode = "ping"
    OpcodePong   Opcode = "pong"
    OpcodeClose  Opcode = "close"
)


