package card

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ebfe/scard"
	"go.uber.org/zap"
)

type EmvData struct {
	Pan             string
	Expiry          string
	Psn             string
	AppId           string
	FingerprintHash string
}

var ppse = []byte("2PAY.SYS.DDF01")

type tlvNode struct {
	tag      uint32
	tagBytes []byte
	length   int
	value    []byte
	children []tlvNode
}

type dolItem struct {
	tag    uint32
	length int
}

type aflEntry struct {
	sfi   int
	first int
	last  int
}

type apduChannel struct {
	card *scard.Card
}

func GetEmvData(card *scard.Card) (*EmvData, error) {
	ch := &apduChannel{
		card: card,
	}

	ppseResp, sw1, sw2, err := selectByName(ch, ppse)
	if err != nil {
		return nil, fmt.Errorf("PPSE SELECT transmit failed: %w", err)
	}
	if !isSWOK(sw1, sw2) {
		return nil, fmt.Errorf("PPSE SELECT failed: SW=%02X%02X (tip: pass --aid <hex>)", sw1, sw2)
	}

	ppseNodes, err := parseTLV(ppseResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PPSE response TLV: %w", err)
	}
	candidateAIDs := dedupeByteSlices(findTagValues(ppseNodes, 0x4F))
	if len(candidateAIDs) == 0 {
		return nil, fmt.Errorf("no AID found")
	}

	for i, aid := range candidateAIDs {
		zap.S().Debugf("Application %d: AID %s", i, strings.ToUpper(hex.EncodeToString(aid)))

		selectData, sw1, sw2, err := selectByName(ch, aid)
		if err != nil {
			zap.S().Debugf("SELECT AID transmit failed: %v\n", err)
			continue
		}
		if !isSWOK(sw1, sw2) {
			zap.S().Debugf("SELECT AID failed: SW=%02X%02X\n", sw1, sw2)
			continue
		}

		var selectNodes []tlvNode
		selectNodes, err = parseTLV(selectData)
		if err != nil {
			zap.S().Debugf("Warning: could not parse SELECT response TLV: %v\n", err)
		}

		var pdol []byte
		if len(selectNodes) > 0 {
			pdolValues := findTagValues(selectNodes, 0x9F38)
			if len(pdolValues) > 0 {
				pdol = pdolValues[0]
			}
		}

		if pdol != nil {
			zap.S().Debugf("PDOL present: %s\n", strings.ToUpper(hex.EncodeToString(pdol)))
		} else {
			zap.S().Debugf("PDOL absent")
		}

		gpoAPDU, err := buildGPOFromPDOL(pdol)
		if err != nil {
			zap.S().Debugf("Failed to build GPO APDU: %v\n", err)
			continue
		}
		gpoData, gsw1, gsw2, err := ch.transmit(gpoAPDU)
		if err != nil {
			zap.S().Debugf("GPO transmit failed: %v\n", err)
			continue
		}
		if !isSWOK(gsw1, gsw2) {
			zap.S().Debugf("GPO failed: SW=%02X%02X\n", gsw1, gsw2)
			continue
		}

		afl := extractAFL(gpoData)
		recordBlobs := make([][]byte, 0)
		if len(afl) == 0 {
			zap.S().Debugf("No AFL returned by GPO")
		} else {
			zap.S().Debugf("AFL: %s\n", strings.ToUpper(hex.EncodeToString(afl)))
			recordBlobs = readRecordsFromAFL(ch, afl)
			zap.S().Debugf("Records read: %d\n", len(recordBlobs))
		}

		extraBlobs := make([][]byte, 0, 1)
		atcResp, atcSW1, atcSW2, err := ch.transmit(mustHex("80CA9F3600"))
		if err == nil && isSWOK(atcSW1, atcSW2) {
			extraBlobs = append(extraBlobs, atcResp)
		}

		wantedTags := []uint32{0x4F, 0x5A, 0x5F34, 0x5F24, 0x57, 0x9F36}
		blobs := make([][]byte, 0, 2+len(recordBlobs)+len(extraBlobs))
		blobs = append(blobs, selectData, gpoData)
		blobs = append(blobs, recordBlobs...)
		blobs = append(blobs, extraBlobs...)
		tagMap := collectTags(blobs, wantedTags)

		var pan string
		if vals := tagMap[0x5A]; len(vals) > 0 {
			pan = parsePANFrom5A(vals[0])
		} else if vals := tagMap[0x57]; len(vals) > 0 {
			pan = parsePANFrom57(vals[0])
		}

		if pan == "" {
			fmt.Println("No PAN found (5A/57). Cannot build EMV fingerprint for this app.")
			continue
		}

		psn := firstOrNil(tagMap[0x5F34])
		exp := "n/a"
		if vals := tagMap[0x5F24]; len(vals) > 0 {
			exp = decodeExpiry5F24(vals[0])
		}
		zap.S().Debugf("EXP: %s", exp)
		atc := "n/a"
		if vals := tagMap[0x9F36]; len(vals) > 0 {
			atc = strings.ToUpper(hex.EncodeToString(vals[0]))
		}
		zap.S().Debugf("ATC: %s", atc)
		psnStr := "n/a"
		if psn != nil {
			psnStr = strings.ToUpper(hex.EncodeToString(psn))
		}
		zap.S().Debugf("PSN: %s", psnStr)

		material := fingerprintMaterial(aid, pan, psn, exp)
		fp := hashFingerprint(material)
		zap.S().Debugf("Fingerprint (sha256): %s\n", fp)
		zap.S().Debugf("Fingerprint material: %s\n", string(material))

		return &EmvData{Pan: pan, Expiry: exp, Psn: psnStr, AppId: strings.ToUpper(hex.EncodeToString(aid)), FingerprintHash: fp}, nil
	}

	return nil, fmt.Errorf("no usable application for this card found")
}

func (ch *apduChannel) transmit(apdu []byte) ([]byte, byte, byte, error) {
	zap.S().Debugf("> %s\n", strings.ToUpper(hex.EncodeToString(apdu)))

	resp, err := ch.card.Transmit(apdu)
	if err != nil {
		return nil, 0, 0, err
	}
	data, sw1, sw2, err := splitResponse(resp)
	if err != nil {
		return nil, 0, 0, err
	}

	if sw1 == 0x6C && len(apdu) >= 5 {
		fixed := append([]byte(nil), apdu...)
		fixed[len(fixed)-1] = sw2
		zap.S().Debugf("< SW=6C%02X; retry LE=%02X\n", sw2, sw2)

		resp, err = ch.card.Transmit(fixed)
		if err != nil {
			return nil, 0, 0, err
		}
		data, sw1, sw2, err = splitResponse(resp)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	if sw1 == 0x61 {
		buf := make([]byte, 0, len(data)+256)
		buf = append(buf, data...)

		for sw1 == 0x61 {
			le := sw2
			getResp := []byte{0x00, 0xC0, 0x00, 0x00, le}
			zap.S().Debugf("> %s\n", strings.ToUpper(hex.EncodeToString(getResp)))

			resp, err = ch.card.Transmit(getResp)
			if err != nil {
				return nil, 0, 0, err
			}
			part, nsw1, nsw2, err := splitResponse(resp)
			if err != nil {
				return nil, 0, 0, err
			}
			buf = append(buf, part...)
			sw1, sw2 = nsw1, nsw2
		}

		data = buf
	}

	zap.S().Debugf("< %s SW=%02X%02X\n", strings.ToUpper(hex.EncodeToString(data)), sw1, sw2)

	return data, sw1, sw2, nil
}

func splitResponse(resp []byte) ([]byte, byte, byte, error) {
	if len(resp) < 2 {
		return nil, 0, 0, fmt.Errorf("APDU response too short: %d", len(resp))
	}
	n := len(resp)
	return resp[:n-2], resp[n-2], resp[n-1], nil
}

func selectByName(ch *apduChannel, name []byte) ([]byte, byte, byte, error) {
	if len(name) > 255 {
		return nil, 0, 0, fmt.Errorf("AID too long: %d", len(name))
	}
	apdu := make([]byte, 0, 6+len(name))
	apdu = append(apdu, 0x00, 0xA4, 0x04, 0x00, byte(len(name)))
	apdu = append(apdu, name...)
	apdu = append(apdu, 0x00)
	return ch.transmit(apdu)
}

func buildGPOFromPDOL(pdol []byte) ([]byte, error) {
	data := make([]byte, 0)
	if len(pdol) > 0 {
		dol, err := parseDOL(pdol)
		if err != nil {
			return nil, err
		}
		for _, it := range dol {
			if it.length < 0 {
				return nil, fmt.Errorf("invalid DOL length for tag %X", it.tag)
			}
			data = append(data, bytes.Repeat([]byte{0x00}, it.length)...)
		}
	}

	if len(data) > 255 {
		return nil, fmt.Errorf("PDOL data too long: %d", len(data))
	}
	pdolData := append([]byte{0x83, byte(len(data))}, data...)
	if len(pdolData) > 255 {
		return nil, fmt.Errorf("GPO command data too long: %d", len(pdolData))
	}

	apdu := make([]byte, 0, 6+len(pdolData))
	apdu = append(apdu, 0x80, 0xA8, 0x00, 0x00, byte(len(pdolData)))
	apdu = append(apdu, pdolData...)
	apdu = append(apdu, 0x00)
	return apdu, nil
}

func extractAFL(gpo []byte) []byte {
	if len(gpo) == 0 {
		return nil
	}

	// Format 1: 80 len [AIP(2) AFL(...)]
	if gpo[0] == 0x80 && len(gpo) >= 2 {
		length := int(gpo[1])
		if 2+length <= len(gpo) && length >= 2 {
			value := gpo[2 : 2+length]
			return append([]byte(nil), value[2:]...)
		}
		return nil
	}

	nodes, err := parseTLV(gpo)
	if err != nil {
		return nil
	}
	afls := findTagValues(nodes, 0x94)
	if len(afls) == 0 {
		return nil
	}
	return append([]byte(nil), afls[0]...)
}

func parseAFL(afl []byte) []aflEntry {
	entries := make([]aflEntry, 0, len(afl)/4)
	for i := 0; i+3 < len(afl); i += 4 {
		sfi := int(afl[i] >> 3)
		first := int(afl[i+1])
		last := int(afl[i+2])
		if sfi == 0 || first == 0 || last < first {
			continue
		}
		entries = append(entries, aflEntry{sfi: sfi, first: first, last: last})
	}
	return entries
}

func readRecordsFromAFL(ch *apduChannel, afl []byte) [][]byte {
	records := make([][]byte, 0)
	for _, entry := range parseAFL(afl) {
		p2 := byte((entry.sfi << 3) | 0x04)
		for rec := entry.first; rec <= entry.last; rec++ {
			apdu := []byte{0x00, 0xB2, byte(rec), p2, 0x00}
			data, sw1, sw2, err := ch.transmit(apdu)
			if err != nil {
				continue
			}
			if isSWOK(sw1, sw2) {
				records = append(records, data)
			}
		}
	}
	return records
}

func parseTLV(buf []byte) ([]tlvNode, error) {
	nodes := make([]tlvNode, 0)
	i := 0
	for i < len(buf) {
		tagBytes, next, err := parseTag(buf, i)
		if err != nil {
			return nil, err
		}
		i = next

		length, next, err := parseLength(buf, i)
		if err != nil {
			return nil, err
		}
		i = next
		if i+length > len(buf) {
			return nil, fmt.Errorf("TLV value exceeds buffer")
		}
		value := append([]byte(nil), buf[i:i+length]...)
		i += length

		n := tlvNode{
			tag:      bytesToUint(tagBytes),
			tagBytes: append([]byte(nil), tagBytes...),
			length:   length,
			value:    value,
		}

		if len(tagBytes) > 0 && (tagBytes[0]&0x20) != 0 {
			children, err := parseTLV(value)
			if err == nil {
				n.children = children
			}
		}

		nodes = append(nodes, n)
	}
	return nodes, nil
}

func parseTag(buf []byte, i int) ([]byte, int, error) {
	if i >= len(buf) {
		return nil, i, fmt.Errorf("unexpected end while parsing tag")
	}
	out := []byte{buf[i]}
	i++
	if out[0]&0x1F == 0x1F {
		for {
			if i >= len(buf) {
				return nil, i, fmt.Errorf("unexpected end in long-form tag")
			}
			b := buf[i]
			i++
			out = append(out, b)
			if b&0x80 == 0 {
				break
			}
		}
	}
	return out, i, nil
}

func parseLength(buf []byte, i int) (int, int, error) {
	if i >= len(buf) {
		return 0, i, fmt.Errorf("unexpected end while parsing length")
	}
	b := buf[i]
	i++
	if b < 0x80 {
		return int(b), i, nil
	}
	if b == 0x80 {
		return 0, i, fmt.Errorf("indefinite BER length is unsupported")
	}
	n := int(b & 0x7F)
	if n == 0 || i+n > len(buf) {
		return 0, i, fmt.Errorf("invalid BER length")
	}
	length := bytesToInt(buf[i : i+n])
	return length, i + n, nil
}

func parseDOL(dol []byte) ([]dolItem, error) {
	out := make([]dolItem, 0)
	i := 0
	for i < len(dol) {
		tagBytes, next, err := parseTag(dol, i)
		if err != nil {
			return nil, err
		}
		i = next
		if i >= len(dol) {
			return nil, fmt.Errorf("broken DOL: missing length")
		}
		length := int(dol[i])
		i++
		out = append(out, dolItem{
			tag:    bytesToUint(tagBytes),
			length: length,
		})
	}
	return out, nil
}

func findTagValues(nodes []tlvNode, tag uint32) [][]byte {
	out := make([][]byte, 0)
	for _, n := range nodes {
		if n.tag == tag {
			out = append(out, append([]byte(nil), n.value...))
		}
		if len(n.children) > 0 {
			out = append(out, findTagValues(n.children, tag)...)
		}
	}
	return out
}

func collectTags(buffers [][]byte, wanted []uint32) map[uint32][][]byte {
	out := make(map[uint32][][]byte, len(wanted))
	for _, tag := range wanted {
		out[tag] = [][]byte{}
	}

	for _, buf := range buffers {
		nodes, err := parseTLV(buf)
		if err != nil {
			continue
		}
		for _, tag := range wanted {
			out[tag] = append(out[tag], findTagValues(nodes, tag)...)
		}
	}
	return out
}

func parsePANFrom5A(v []byte) string {
	return strings.TrimRight(bcdToDigits(v), "F")
}

func parsePANFrom57(v []byte) string {
	s := bcdToDigits(v)
	if idx := strings.IndexByte(s, 'D'); idx >= 0 {
		return s[:idx]
	}
	if idx := strings.IndexByte(s, '='); idx >= 0 {
		return s[:idx]
	}
	return ""
}

func decodeExpiry5F24(v []byte) string {
	s := bcdToDigits(v)
	if len(s) >= 6 {
		return fmt.Sprintf("20%s-%s-%s", s[0:2], s[2:4], s[4:6])
	}
	return s
}

func bcdToDigits(v []byte) string {
	return strings.ToUpper(hex.EncodeToString(v))
}

func maskPAN(pan string) string {
	if len(pan) <= 10 {
		return pan
	}
	return pan[:6] + strings.Repeat("*", len(pan)-10) + pan[len(pan)-4:]
}
func fingerprintMaterial(aid []byte, pan string, psn []byte, exp string) []byte {
	psnHex := ""
	if psn != nil {
		psnHex = hex.EncodeToString(psn)
	}
	material := fmt.Sprintf("AID=%s|PAN=%s|PSN=%s|EXP=%s", hex.EncodeToString(aid), pan, psnHex, exp)
	return []byte(material)
}

func hashFingerprint(material []byte) (digest string) {
	sum := sha256.Sum256(material)
	return hex.EncodeToString(sum[:])
}

func isSWOK(sw1 byte, sw2 byte) bool {
	return sw1 == 0x90 && sw2 == 0x00
}

func firstOrNil(values [][]byte) []byte {
	if len(values) == 0 {
		return nil
	}
	return values[0]
}

func bytesToInt(b []byte) int {
	n := 0
	for _, x := range b {
		n = (n << 8) | int(x)
	}
	return n
}

func bytesToUint(b []byte) uint32 {
	var n uint32
	for _, x := range b {
		n = (n << 8) | uint32(x)
	}
	return n
}

func dedupeByteSlices(values [][]byte) [][]byte {
	out := make([][]byte, 0, len(values))
	seen := map[string]struct{}{}
	for _, v := range values {
		k := string(v)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, v)
	}
	return out
}

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}
