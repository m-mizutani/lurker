package usecase

import "fmt"

func (x *Usecase) Loop() error {
	for pkt := range x.clients.Device().ReadPacket() {
		if pkt == nil {
			break
		}
		fmt.Println(pkt)
	}

	return nil
}
