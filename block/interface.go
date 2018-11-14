package block

type ChangeToByte interface{

}

type Hash interface {
	Hash(b [32]byte) error
}

