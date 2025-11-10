package card

import "github.com/ebfe/scard"
import "fmt"
import "paSKUDa/internal/config"
import "paSKUDa/internal/models"
import "context"
import "go.uber.org/zap"

type CardManager struct {
	ctx      context.Context
	scardCtx *scard.Context
	conf     *config.Config
	openChan chan models.DoorSignal
}

func Init(conf *config.Config, ctx context.Context, openChan chan models.DoorSignal) (*CardManager, error) {
	scardCtx, err := scard.EstablishContext()
	if err != nil {
		return nil, err
	}
	zap.S().Infof("init done")
	return &CardManager{conf: conf, ctx: ctx, scardCtx: scardCtx, openChan: openChan}, nil
}

func (cm *CardManager) waitUntilCardPresent(readers []string) (int, error) {
	rs := make([]scard.ReaderState, len(readers))
	for i := range rs {
		rs[i].Reader = readers[i]
		rs[i].CurrentState = scard.StateUnaware
	}

	for {
		select {
		case <-cm.ctx.Done():
			return -1, nil
		default:
		}
		for i := range rs {
			if rs[i].EventState&scard.StatePresent != 0 {
				return i, nil
			}
			rs[i].CurrentState = rs[i].EventState
		}
		err := cm.scardCtx.GetStatusChange(rs, -1)
		if err != nil {
			return -1, err
		}
	}
}

func (cm *CardManager) processCard(reader string) error {
	zap.S().Infof("card found, connecting")
	card, err := cm.scardCtx.Connect(reader, scard.ShareExclusive, scard.ProtocolAny)
	if err != nil {
		return err
	}
	defer card.Disconnect(scard.ResetCard)
	zap.S().Infof("card connected, reading status")

	status, err := card.Status()
	if err != nil {
		return err
	}
	zap.S().Infof("card status: `%x`, active protocol: `%x`, atr: `% x`", status.State, status.ActiveProtocol, status.Atr)
	return nil
}

func (cm *CardManager) Start() error {
	readers, err := cm.scardCtx.ListReaders()
	if err != nil {
		return err
	}
	if len(readers) != 1 {
		return fmt.Errorf("one and only one reader need to be attached.. readers: `%w`", readers)
	}

	for {
		select {
		case <-cm.ctx.Done():
			return nil
		default:
		}
		index, err := cm.waitUntilCardPresent(readers)
		if err != nil {
			zap.S().Errorf("error waiting card: `%w`", err)
			continue
		}
		cm.processCard(readers[index])
	}

	return nil
}
func (cm *CardManager) Deinit() {
	cm.scardCtx.Release()
}
