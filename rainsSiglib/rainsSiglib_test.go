package rainsSiglib

import (
	"net"
	"rains/rainslib"
	"rains/utils/testUtil"
	"rains/utils/zoneFileParser"
	"testing"
	"time"

	"golang.org/x/crypto/ed25519"
)

func TestEncodeAndDecode(t *testing.T) {
	nameObjectContent := rainslib.NameObject{
		Name:  "ethz2.ch",
		Types: []rainslib.ObjectType{rainslib.OTIP4Addr, rainslib.OTIP6Addr},
	}

	publicKey := rainslib.PublicKey{
		KeySpace:   rainslib.RainsKeySpace,
		Type:       rainslib.Ed25519,
		Key:        ed25519.PublicKey([]byte("01234567890123456789012345678901")),
		ValidSince: 10000,
		ValidUntil: 50000,
	}
	certificate := rainslib.CertificateObject{
		Type:     rainslib.PTTLS,
		HashAlgo: rainslib.Sha256,
		Usage:    rainslib.CUEndEntity,
		Data:     []byte("certData"),
	}
	serviceInfo := rainslib.ServiceInfo{
		Name:     "lookup",
		Port:     49830,
		Priority: 1,
	}

	nameObject := rainslib.Object{Type: rainslib.OTName, Value: nameObjectContent}
	ip6Object := rainslib.Object{Type: rainslib.OTIP6Addr, Value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}
	ip4Object := rainslib.Object{Type: rainslib.OTIP4Addr, Value: "127.0.0.1"}
	redirObject := rainslib.Object{Type: rainslib.OTRedirection, Value: "ns.ethz.ch"}
	delegObject := rainslib.Object{Type: rainslib.OTDelegation, Value: publicKey}
	nameSetObject := rainslib.Object{Type: rainslib.OTNameset, Value: rainslib.NamesetExpression("Would be an expression")}
	certObject := rainslib.Object{Type: rainslib.OTCertInfo, Value: certificate}
	serviceInfoObject := rainslib.Object{Type: rainslib.OTServiceInfo, Value: serviceInfo}
	registrarObject := rainslib.Object{Type: rainslib.OTRegistrar, Value: "Registrar information"}
	registrantObject := rainslib.Object{Type: rainslib.OTRegistrant, Value: "Registrant information"}
	infraObject := rainslib.Object{Type: rainslib.OTInfraKey, Value: publicKey}
	extraObject := rainslib.Object{Type: rainslib.OTExtraKey, Value: publicKey}
	nextKey := rainslib.Object{Type: rainslib.OTNextKey, Value: publicKey}

	signature := rainslib.Signature{
		KeySpace:   rainslib.RainsKeySpace,
		Algorithm:  rainslib.Ed25519,
		ValidSince: time.Now().Unix(),
		ValidUntil: time.Now().Add(24 * time.Hour).Unix(),
	}

	_, subjectAddress1, _ := net.ParseCIDR("127.0.0.1/32")
	_, subjectAddress2, _ := net.ParseCIDR("127.0.0.1/24")
	_, subjectAddress3, _ := net.ParseCIDR("2001:db8::/32")

	assertion := &rainslib.AssertionSection{
		Content: []rainslib.Object{nameObject, ip6Object, ip4Object, redirObject, delegObject, nameSetObject, certObject, serviceInfoObject, registrarObject,
			registrantObject, infraObject, extraObject, nextKey},
		Context:     ".",
		SubjectName: "ethz",
		SubjectZone: "ch",
	}

	shard := &rainslib.ShardSection{
		Content:     []*rainslib.AssertionSection{assertion},
		Context:     ".",
		SubjectZone: "ch",
		RangeFrom:   "aaa",
		RangeTo:     "zzz",
	}

	zone := &rainslib.ZoneSection{
		Content:     []rainslib.MessageSectionWithSig{assertion, shard},
		Context:     ".",
		SubjectZone: "ch",
	}

	query := &rainslib.QuerySection{
		Context: ".",
		Expires: 159159,
		Name:    "ethz.ch",
		Options: []rainslib.QueryOption{rainslib.QOMinE2ELatency, rainslib.QOMinInfoLeakage},
		Token:   rainslib.GenerateToken(),
		Type:    rainslib.OTIP4Addr,
	}

	notification := &rainslib.NotificationSection{
		Token: rainslib.GenerateToken(),
		Type:  rainslib.NoAssertionsExist,
		Data:  "Notification information",
	}

	addressAssertion1 := &rainslib.AddressAssertionSection{
		SubjectAddr: subjectAddress1,
		Context:     ".",
		Content:     []rainslib.Object{nameObject},
	}

	addressAssertion2 := &rainslib.AddressAssertionSection{
		SubjectAddr: subjectAddress2,
		Context:     ".",
		Content:     []rainslib.Object{redirObject, delegObject, registrantObject},
	}

	addressAssertion3 := &rainslib.AddressAssertionSection{
		SubjectAddr: subjectAddress3,
		Context:     ".",
		Content:     []rainslib.Object{redirObject, delegObject, registrantObject},
	}

	addressZone := &rainslib.AddressZoneSection{
		SubjectAddr: subjectAddress2,
		Context:     ".",
		Content:     []*rainslib.AddressAssertionSection{addressAssertion1, addressAssertion2, addressAssertion3},
	}

	addressQuery := &rainslib.AddressQuerySection{
		SubjectAddr: subjectAddress1,
		Context:     ".",
		Expires:     7564859,
		Token:       rainslib.GenerateToken(),
		Type:        rainslib.OTName,
		Options:     []rainslib.QueryOption{rainslib.QOMinE2ELatency, rainslib.QOMinInfoLeakage},
	}

	message := rainslib.RainsMessage{
		Content: []rainslib.MessageSection{
			assertion,
			shard,
			zone,
			query,
			notification,
			addressAssertion1,
			addressAssertion2,
			addressAssertion3,
			addressZone,
			addressQuery,
		},
		Token:        rainslib.GenerateToken(),
		Capabilities: []rainslib.Capability{rainslib.Capability("Test"), rainslib.Capability("Yes!")},
	}

	genPublicKey, genPrivateKey, _ := ed25519.GenerateKey(nil)
	pKey := rainslib.PublicKey{
		KeySpace:   rainslib.RainsKeySpace,
		Type:       rainslib.Ed25519,
		ValidSince: time.Now().Add(-24 * time.Hour).Unix(),
		ValidUntil: time.Now().Add(24 * time.Hour).Unix(),
		Key:        genPublicKey,
	}
	pKeys := make(map[rainslib.KeyAlgorithmType]rainslib.PublicKey)
	pKeys[rainslib.KeyAlgorithmType(pKey.Type)] = pKey
	maxValidity := rainslib.MaxCacheValidity{
		AssertionValidity:        30 * time.Hour,
		ShardValidity:            30 * time.Hour,
		ZoneValidity:             30 * time.Hour,
		AddressAssertionValidity: 30 * time.Hour,
		AddressZoneValidity:      30 * time.Hour,
	}
	ok := SignMessage(&message, genPrivateKey, signature, zoneFileParser.Parser{})
	if !ok {
		t.Error("Was not able to generate and add a signature to the message")
	}
	ok = CheckMessageSignatures(&message, pKey, zoneFileParser.Parser{}, maxValidity)
	if !ok {
		t.Error("Verification of message signature failed")
	}

	ok = SignSection(assertion, genPrivateKey, signature, zoneFileParser.Parser{})
	if !ok {
		t.Error("Was not able to generate and add a signature to the assertion")
	}
	ok = CheckSectionSignatures(assertion, pKeys, zoneFileParser.Parser{}, maxValidity)
	if !ok {
		t.Error("Verification of assertion signature failed")
	}

	ok = SignSection(shard, genPrivateKey, signature, zoneFileParser.Parser{})
	if !ok {
		t.Error("Was not able to generate and add a signature to the shard")
	}
	ok = CheckSectionSignatures(shard, pKeys, zoneFileParser.Parser{}, maxValidity)
	if !ok {
		t.Error("Verification of shard signature failed")
	}

	ok = SignSection(zone, genPrivateKey, signature, zoneFileParser.Parser{})
	if !ok {
		t.Error("Was not able to generate and add a signature to the zone")
	}
	ok = CheckSectionSignatures(zone, pKeys, zoneFileParser.Parser{}, maxValidity)
	if !ok {
		t.Error("Verification of zone signature failed")
	}

	ok = SignSection(addressAssertion1, genPrivateKey, signature, zoneFileParser.Parser{})
	if !ok {
		t.Error("Was not able to generate and add a signature to the addressAssertion")
	}
	ok = CheckSectionSignatures(addressAssertion1, pKeys, zoneFileParser.Parser{}, maxValidity)
	if !ok {
		t.Error("Verification of addressAssertion signature failed")
	}

	ok = SignSection(addressAssertion2, genPrivateKey, signature, zoneFileParser.Parser{})
	if !ok {
		t.Error("Was not able to generate and add a signature to the addressAssertion")
	}
	ok = CheckSectionSignatures(addressAssertion2, pKeys, zoneFileParser.Parser{}, maxValidity)
	if !ok {
		t.Error("Verification of addressAssertion signature failed")
	}

	ok = SignSection(addressAssertion3, genPrivateKey, signature, zoneFileParser.Parser{})
	if !ok {
		t.Error("Was not able to generate and add a signature to the addressAssertion")
	}
	ok = CheckSectionSignatures(addressAssertion3, pKeys, zoneFileParser.Parser{}, maxValidity)
	if !ok {
		t.Error("Verification of addressAssertion signature failed")
	}

	ok = SignSection(addressZone, genPrivateKey, signature, zoneFileParser.Parser{})
	if !ok {
		t.Error("Was not able to generate and add a signature to the addressZone")
	}
	ok = CheckSectionSignatures(addressZone, pKeys, zoneFileParser.Parser{}, maxValidity)
	if !ok {
		t.Error("Verification of addressZone signature failed")
	}
}

func TestSignMessageErrors(t *testing.T) {

}

func TestCheckMessageStringFields(t *testing.T) {
	message := testUtil.GetMessage()
	var tests = []struct {
		input *rainslib.RainsMessage
		want  bool
	}{
		{nil, false},
		{&message, true},
		{&rainslib.RainsMessage{Capabilities: []rainslib.Capability{rainslib.Capability(":ip:")}}, false},
		{&rainslib.RainsMessage{Content: []rainslib.MessageSection{&rainslib.AssertionSection{SubjectName: ":ip:"}}}, false},
	}
	for _, test := range tests {
		if checkMessageStringFields(test.input) != test.want {
			t.Errorf("expected=%v, actual=%v, value=%v", test.want, checkMessageStringFields(test.input), test.input)
		}
	}
}

func TestCheckStringFields(t *testing.T) {
	sections := testUtil.GetMessage().Content
	var tests = []struct {
		input rainslib.MessageSection
		want  bool
	}{
		{nil, false},
		{sections[0], true},
		{sections[1], true},
		{sections[2], true},
		{sections[3], true},
		{sections[4], true},
		{sections[5], true},
		{sections[6], true},
		{sections[8], true},
		{sections[9], true},
		{&rainslib.AssertionSection{SubjectName: ":ip:"}, false},
		{&rainslib.AssertionSection{Context: ":ip:"}, false},
		{&rainslib.AssertionSection{SubjectZone: ":ip:"}, false},
		{&rainslib.AssertionSection{Content: []rainslib.Object{rainslib.Object{Type: rainslib.OTRegistrar, Value: ":ip55:"}}}, false},
		{&rainslib.ShardSection{Context: ":ip:"}, false},
		{&rainslib.ShardSection{SubjectZone: ":ip:"}, false},
		{&rainslib.ShardSection{RangeFrom: ":ip:"}, false},
		{&rainslib.ShardSection{RangeTo: ":ip:"}, false},
		{&rainslib.ShardSection{Content: []*rainslib.AssertionSection{&rainslib.AssertionSection{SubjectName: ":ip:"}}}, false},
		{&rainslib.ZoneSection{SubjectZone: ":ip:"}, false},
		{&rainslib.ZoneSection{Context: ":ip:"}, false},
		{&rainslib.ZoneSection{Content: []rainslib.MessageSectionWithSig{&rainslib.AssertionSection{SubjectName: ":ip:"}}}, false},
		{&rainslib.QuerySection{Context: ":ip:"}, false},
		{&rainslib.QuerySection{Name: ":ip:"}, false},
		{&rainslib.NotificationSection{Data: ":ip:"}, false},
		{&rainslib.AddressQuerySection{Context: ":ip:"}, false},
		{&rainslib.AddressAssertionSection{Context: ":ip:"}, false},
		{&rainslib.AddressAssertionSection{Content: []rainslib.Object{rainslib.Object{Type: rainslib.OTRegistrant, Value: ":ip55:"}}}, false},
		{&rainslib.AddressZoneSection{Context: ":ip:"}, false},
		{&rainslib.AddressZoneSection{Content: []*rainslib.AddressAssertionSection{&rainslib.AddressAssertionSection{Context: ":ip:"}}}, false},
	}
	for _, test := range tests {
		if checkStringFields(test.input) != test.want {
			t.Errorf("expected=%v, actual=%v, value=%v", test.want, checkStringFields(test.input), test.input)
		}
	}
}

func TestCheckObjectFields(t *testing.T) {
	var tests = []struct {
		input []rainslib.Object
		want  bool
	}{
		{nil, true},
		{[]rainslib.Object{}, true},
		{testUtil.GetAllValidObjects(), true},
		{[]rainslib.Object{rainslib.Object{Type: rainslib.OTName, Value: rainslib.NameObject{Name: ":ip55:"}}}, false},
		{[]rainslib.Object{rainslib.Object{Type: rainslib.OTRedirection, Value: ":ip55:"}}, false},
		{[]rainslib.Object{rainslib.Object{Type: rainslib.OTNameset, Value: rainslib.NamesetExpression(":ip55:")}}, false},
		{[]rainslib.Object{rainslib.Object{Type: rainslib.OTServiceInfo, Value: rainslib.ServiceInfo{Name: ":ip55:"}}}, false},
		{[]rainslib.Object{rainslib.Object{Type: rainslib.OTRegistrar, Value: ":ip55:"}}, false},
		{[]rainslib.Object{rainslib.Object{Type: rainslib.OTRegistrant, Value: ":ip55:"}}, false},
		{[]rainslib.Object{rainslib.Object{Type: rainslib.ObjectType(-1), Value: nil}}, false},
	}
	for _, test := range tests {
		if checkObjectFields(test.input) != test.want {
			t.Errorf("expected=%v, actual=%v, value=%v", test.want, checkObjectFields(test.input), test.input)
		}
	}
}

func TestCheckCapabilites(t *testing.T) {
	var tests = []struct {
		input []rainslib.Capability
		want  bool
	}{
		{nil, true},
		{[]rainslib.Capability{}, true},
		{[]rainslib.Capability{rainslib.Capability("")}, true},
		{[]rainslib.Capability{rainslib.Capability("Good")}, true},
		{[]rainslib.Capability{rainslib.Capability(":ip: bad")}, false},
		{[]rainslib.Capability{rainslib.Capability(":ip:")}, false},
		{[]rainslib.Capability{rainslib.Capability("bad :ip: test")}, false},
		{[]rainslib.Capability{rainslib.Capability("bad\t:ip:\ttest")}, false},
		{[]rainslib.Capability{rainslib.Capability("bad\n:ip:\ntest")}, false},
		{[]rainslib.Capability{rainslib.Capability("bad\n:ip:\ttest")}, false},
		{[]rainslib.Capability{rainslib.Capability("bad test :ip:")}, false},
		{[]rainslib.Capability{rainslib.Capability("bad test\n\n :ip:")}, false},
		{[]rainslib.Capability{rainslib.Capability("as:Good:dh")}, true},
		{[]rainslib.Capability{rainslib.Capability("as:Good: dh")}, true},
		{[]rainslib.Capability{rainslib.Capability("as :Good:dh")}, true},
		{[]rainslib.Capability{rainslib.Capability(" :: ")}, true},
		{[]rainslib.Capability{rainslib.Capability("::")}, true},
		{[]rainslib.Capability{rainslib.Capability("::"), rainslib.Capability(":ip4:Good")}, true},
		{[]rainslib.Capability{rainslib.Capability("::"), rainslib.Capability(":ip4: Good")}, false},
	}
	for _, test := range tests {
		if checkCapabilites(test.input) != test.want {
			t.Errorf("expected=%v, actual=%v, value=%v", test.want, checkCapabilites(test.input), test.input)
		}
	}
}

func TestContainsZoneFileType(t *testing.T) {
	var tests = []struct {
		input string
		want  bool
	}{
		{"", false},
		{"Good", false},
		{":ip:", true},
		{":ip: bad", true},
		{"bad test\n\n :ip:", true},
		{"bad :ip: test", true},
		{"bad\t:ip:\ttest", true},
		{"bad\n:ip:\ntest", true},
		{"bad\n:ip:\ttest", true},
		{"bad test :ip:", true},
		{"as:Good:dh", false},
		{"as:Good: dh", false},
		{"as :Good:dh", false},
		{":ip:d", false},
		{" :: ", false},
		{"::", false},
	}
	for _, test := range tests {
		if containsZoneFileType(test.input) != test.want {
			t.Errorf("expected=%v, actual=%v, value=%v", test.want, containsZoneFileType(test.input), test.input)
		}
	}
}
