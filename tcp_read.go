package endless_tcp

type ReadMes struct {
	N   int
	Mes []byte
}

type UpgradeRead interface {
	ReadMessage(b *ReadMes)
}
