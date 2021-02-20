package protocol

const ()

//ArpHeader 总共28字节
type ArpHeader struct {
	HardType      [2]byte
	Prot          [2]byte
	HardAddrLen   [1]byte
	ProtLen       [1]byte
	OperationType [2]byte
	SenderMac     [6]byte
	SenderIP      [4]byte
	RecverMac     [6]byte
	RecverIP      [4]byte
}

//ParseArpHeader 解析arp数据包
func ParseArpHeader(buf []byte) *ArpHeader {
	var hardType [2]byte
	var prot [2]byte
	var hardAddrLen [1]byte
	var protLen [1]byte
	var operationType [2]byte
	var senderMac [6]byte
	var senderIP [4]byte
	var recverMac [6]byte
	var recverIP [4]byte

	copy(hardType[:], buf[:2])
	copy(prot[:], buf[2:4])
	copy(hardAddrLen[:], buf[4:5])
	copy(protLen[:], buf[5:6])
	copy(operationType[:], buf[6:8])
	copy(senderMac[:], buf[8:14])
	copy(senderIP[:], buf[14:18])
	copy(recverMac[:], buf[18:24])
	copy(recverIP[:], buf[24:28])
	arph := &ArpHeader{
		HardType:      hardType,
		Prot:          prot,
		HardAddrLen:   hardAddrLen,
		ProtLen:       protLen,
		OperationType: operationType,
		SenderMac:     senderMac,
		SenderIP:      senderIP,
		RecverMac:     recverMac,
		RecverIP:      recverIP,
	}
	return arph
}

//ToBytes header to []byte
func (ah *ArpHeader) ToBytes() []byte {
	ret := make([]byte, 28)
	copy(ret[0:2], ah.HardType[:])
	copy(ret[2:4], ah.Prot[:])
	copy(ret[4:5], ah.HardAddrLen[:])
	copy(ret[5:6], ah.ProtLen[:])
	copy(ret[6:8], ah.OperationType[:])
	copy(ret[8:14], ah.SenderMac[:])
	copy(ret[14:18], ah.SenderIP[:])
	copy(ret[18:24], ah.RecverMac[:])
	copy(ret[24:28], ah.RecverIP[:])
	return ret
}
