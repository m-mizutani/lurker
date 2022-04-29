package usecase

import "fmt"

func (x *Usecase) Loop() error {
	fmt.Println("start monitoring...")

	for pkt := range x.clients.Device().ReadPacket() {
		if pkt == nil {
			break
		}

		for _, hdlr := range x.handlers {
			if err := hdlr.Handle(pkt, x.output); err != nil {
				return err
			}
		}
	}

	return nil
}
