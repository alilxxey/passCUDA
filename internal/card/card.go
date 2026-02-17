package card

import "github.com/ebfe/scard"
import "fmt"
import "paSKUDa/internal/config"
import "paSKUDa/internal/models"
import "paSKUDa/internal/door"
import "slices"
import "context"
import "go.uber.org/zap"

type CardManager struct {
	ctx              context.Context
	scardCtx         *scard.Context
	conf             *config.Config
	openChan         chan models.DoorSignal
	adminMessageChan chan string
	door             *door.Door
	tempBanList      []string
}

func Init(conf *config.Config, ctx context.Context, openChan chan models.DoorSignal, adminMessageChan chan string, door *door.Door) (*CardManager, error) {
	scardCtx, err := scard.EstablishContext()
	if err != nil {
		return nil, err
	}
	zap.S().Infof("init done")
	return &CardManager{conf: conf, ctx: ctx, scardCtx: scardCtx, openChan: openChan, adminMessageChan: adminMessageChan, door: door}, nil
}

func (cm *CardManager) selectReader() (string, error) {
	readers, err := cm.scardCtx.ListReaders()
	if err != nil {
		return "", fmt.Errorf("failed to list readers: %w", err)
	}
	if len(readers) == 0 {
		return "", fmt.Errorf("no PC/SC readers found")
	}
	if len(readers) > 1 {
		return "", fmt.Errorf("one and only ony PC/SC reader can be attached")
	}
	return readers[0], nil
}

func (cm *CardManager) waitForCard(rs []scard.ReaderState) error {
	for {
		if err := cm.ctx.Err(); err != nil {
			return err
		}
		if err := cm.scardCtx.GetStatusChange(rs, -1); err != nil {
			return err
		}
		st := rs[0].EventState
		rs[0].CurrentState = st &^ scard.StateChanged

		// if card is present and was not before
		if st&scard.StatePresent != 0 {
			return nil
		}
	}
}

func (cm *CardManager) selectCard(readerName string) (*scard.Card, error) {
	card, err := cm.scardCtx.Connect(readerName, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (cm *CardManager) banUserAndReport(fingerprint string, msg string) {
	select {
	case cm.adminMessageChan <- msg:
	default:
	}
	zap.S().Warn(msg)
	cm.tempBanList = append(cm.tempBanList, fingerprint)
	zap.S().Warnf("user with EMV card fingerprint: `%s` is banned", fingerprint)
}

func (cm *CardManager) openDoorAndReport(fingerprint string, user *config.User) {
	select {
	case cm.adminMessageChan <- fmt.Sprintf("door opened by EMV card with fingerprint: `%v`, by user: `%v`", fingerprint, user):
	default:
	}
	select {
	case cm.openChan <- models.DoorSignalOpen:
	default:
	}
}

func (cm *CardManager) processCard(card *scard.Card) error {
	emvData, err := GetEmvData(card)
	if err != nil {
		return err
	}
	zap.S().Infof("card data: %s\n", emvData)
	if slices.Contains(cm.tempBanList, emvData.FingerprintHash) {
		return nil
	}

	user, err := cm.conf.FindUserByCardFingerprint(emvData.FingerprintHash)
	if err != nil {
		msg := fmt.Sprintf("unregistred EMV card detected, fingerprint: %s, card is banned", emvData.FingerprintHash)
		cm.banUserAndReport(emvData.FingerprintHash, msg)
		return err
	}
	if cm.door.OpenInProgress {
		zap.S().Warnf("user: `%s` trying to open door while previous open in progress", user)
		return nil
	}
	cm.openDoorAndReport(emvData.FingerprintHash, user)

	return nil
}

func (cm *CardManager) Start() error {
	readerName, err := cm.selectReader()
	if err != nil {
		return fmt.Errorf("failed to select reader: %v", err)
	}
	zap.S().Infof("using reader: %s", readerName)
	go func() {
		rs := []scard.ReaderState{
			{
				Reader:       readerName,
				CurrentState: scard.StateUnaware,
			},
		}
		for {
			if err := cm.ctx.Err(); err != nil {
				return
			}
			err := cm.waitForCard(rs)
			if err != nil {
				zap.S().Errorf("failed while waiting for card: %v", err)
			}
			zap.S().Infof("card detected")
			card, err := cm.selectCard(readerName)
			if err != nil {
				zap.S().Errorf("failed for select card: %v", err)
				continue
			}
			zap.S().Debugf("card selected: %v\n", card)
			if err := cm.processCard(card); err != nil {
				zap.S().Errorf("failed to process card: %v", err)
			}
			_ = card.Disconnect(scard.LeaveCard)
		}
	}()
	return nil
}
func (cm *CardManager) Deinit() {
	cm.scardCtx.Release()
}
