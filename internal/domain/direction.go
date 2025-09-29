package domain

type Direction string

const (
    DirectionClientToUpstream Direction = "client->upstream"
    DirectionUpstreamToClient Direction = "upstream->client"
)


