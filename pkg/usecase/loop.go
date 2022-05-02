package usecase

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (x *Usecase) Loop() error {
	fmt.Println("start monitoring...")

	ticker := time.NewTicker(time.Second)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

mainLoop:
	for {
		select {
		case pkt := <-x.clients.Device().ReadPacket():
			if pkt == nil {
				break mainLoop
			}

			for _, hdlr := range x.handlers {
				if err := hdlr.Handle(pkt, x.spouts); err != nil {
					return err
				}
			}

		case <-ticker.C:
			for _, hdlr := range x.handlers {
				if err := hdlr.Tick(x.spouts); err != nil {
					return err
				}
			}

		case s := <-sigCh:
			fmt.Printf("\nShutting down by %s\n", s)
			break mainLoop
		}
	}

	return nil
}
