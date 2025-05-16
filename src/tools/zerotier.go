package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

const (
	ZT_C25519_PUBLIC_KEY_LEN = 64
	ZT_C25519_SIGNATURE_LEN  = 96
	ZT_WORLD_MAX_ROOTS       = 4
	ZT_WORLD_ID_EARTH        = 149604618
	ZT_INETADDRESS_IPV4      = 0x04
	ZT_INETADDRESS_IPV6      = 0x06
)

type World struct {
	Type                  uint8
	ID                    uint64
	Timestamp             uint64
	UpdatesMustBeSignedBy [ZT_C25519_PUBLIC_KEY_LEN]byte
	Signature             [ZT_C25519_SIGNATURE_LEN]byte
	Roots                 []Root
	RawData               []byte
}

type Root struct {
	Identity        Identity
	StableEndpoints []InetAddress
}

type Identity struct {
	Address   [5]byte
	PublicKey [64]byte
}

type InetAddress struct {
	Family uint8
	IP     net.IP
	Port   uint16
}

func ParseWorld(data []byte) (*World, error) {
	if len(data) < 1+8+8+ZT_C25519_PUBLIC_KEY_LEN+ZT_C25519_SIGNATURE_LEN+1 {
		return nil, fmt.Errorf("data too short to be a valid world file")
	}

	buf := bytes.NewReader(data)
	world := &World{
		RawData: data,
	}

	// Read type (1 byte)
	if err := binary.Read(buf, binary.BigEndian, &world.Type); err != nil {
		return nil, fmt.Errorf("reading type: %w", err)
	}

	// Read ID (8 bytes)
	if err := binary.Read(buf, binary.BigEndian, &world.ID); err != nil {
		return nil, fmt.Errorf("reading id: %w", err)
	}

	// Read timestamp (8 bytes)
	if err := binary.Read(buf, binary.BigEndian, &world.Timestamp); err != nil {
		return nil, fmt.Errorf("reading timestamp: %w", err)
	}

	// Read update signer public key
	if _, err := buf.Read(world.UpdatesMustBeSignedBy[:]); err != nil {
		return nil, fmt.Errorf("reading update signer public key: %w", err)
	}

	// Read signature
	if _, err := buf.Read(world.Signature[:]); err != nil {
		return nil, fmt.Errorf("reading signature: %w", err)
	}

	// Read number of roots (1 byte)
	var numRoots uint8
	if err := binary.Read(buf, binary.BigEndian, &numRoots); err != nil {
		return nil, fmt.Errorf("reading number of roots: %w", err)
	}

	if numRoots > ZT_WORLD_MAX_ROOTS {
		return nil, fmt.Errorf("too many roots (%d > max %d)", numRoots, ZT_WORLD_MAX_ROOTS)
	}

	// Parse each root
	for i := 0; i < int(numRoots); i++ {
		root, err := ParseRoot(buf)
		if err != nil {
			return nil, fmt.Errorf("parsing root %d: %w", i, err)
		}
		world.Roots = append(world.Roots, *root)
	}

	return world, nil
}

func ParseRoot(buf *bytes.Reader) (*Root, error) {
	root := &Root{}

	// Parse identity
	identity, err := ParseIdentity(buf)
	if err != nil {
		return nil, fmt.Errorf("parsing identity: %w", err)
	}
	root.Identity = *identity

	// Read number of endpoints (1 byte)
	var numEndpoints uint8
	if err := binary.Read(buf, binary.BigEndian, &numEndpoints); err != nil {
		return nil, fmt.Errorf("reading number of endpoints: %w", err)
	}

	// Parse each endpoint
	for i := 0; i < int(numEndpoints); i++ {
		ep, err := ParseInetAddress(buf)
		if err != nil {
			return nil, fmt.Errorf("parsing endpoint %d: %w", i, err)
		}
		root.StableEndpoints = append(root.StableEndpoints, *ep)
	}

	return root, nil
}

func ParseIdentity(buf *bytes.Reader) (*Identity, error) {
	identity := &Identity{}

	// Read address (5 bytes)
	if _, err := buf.Read(identity.Address[:]); err != nil {
		return nil, fmt.Errorf("reading address: %w", err)
	}

	// Read identity type (1 byte)
	var identityType uint8
	if err := binary.Read(buf, binary.BigEndian, &identityType); err != nil {
		return nil, fmt.Errorf("reading identity type: %w", err)
	}
	if identityType != 0 {
		return nil, fmt.Errorf("unsupported identity type %d (only 0=C25519/Ed25519 is supported)", identityType)
	}

	// Read public key (64 bytes)
	if _, err := buf.Read(identity.PublicKey[:]); err != nil {
		return nil, fmt.Errorf("reading public key: %w", err)
	}

	// Skip private key if present (we don't need it)
	var privateKeyLen uint8
	if err := binary.Read(buf, binary.BigEndian, &privateKeyLen); err != nil {
		return nil, fmt.Errorf("reading private key length: %w", err)
	}
	if privateKeyLen > 0 {
		if _, err := buf.Seek(int64(privateKeyLen), 1); err != nil {
			return nil, fmt.Errorf("skipping private key: %w", err)
		}
	}

	return identity, nil
}

func ParseInetAddress(buf *bytes.Reader) (*InetAddress, error) {
	addr := &InetAddress{}

	// Read address family (1 byte)
	if err := binary.Read(buf, binary.BigEndian, &addr.Family); err != nil {
		return nil, fmt.Errorf("reading address family: %w", err)
	}

	switch addr.Family {
	case ZT_INETADDRESS_IPV4:
		var ip [4]byte
		if _, err := buf.Read(ip[:]); err != nil {
			return nil, fmt.Errorf("reading IPv4 address: %w", err)
		}
		addr.IP = net.IPv4(ip[0], ip[1], ip[2], ip[3])
	case ZT_INETADDRESS_IPV6:
		var ip [16]byte
		if _, err := buf.Read(ip[:]); err != nil {
			return nil, fmt.Errorf("reading IPv6 address: %w", err)
		}
		addr.IP = net.IP(ip[:])
	default:
		return nil, fmt.Errorf("unsupported address family %d", addr.Family)
	}

	// Read port (2 bytes)
	if err := binary.Read(buf, binary.BigEndian, &addr.Port); err != nil {
		return nil, fmt.Errorf("reading port: %w", err)
	}

	return addr, nil
}

func (i *Identity) String() string {
	return fmt.Sprintf("%s:%s", hex.EncodeToString(i.Address[:]), hex.EncodeToString(i.PublicKey[:]))
}

func (ia *InetAddress) String() string {
	if ia.IP == nil {
		return "invalid"
	}

	ipStr := ia.IP.String()
	if ia.Port != 0 {
		return fmt.Sprintf("%s:%d", ipStr, ia.Port)
	}
	return ipStr
}

func (w World) ToBase64() string {
	return base64.StdEncoding.EncodeToString(w.RawData)
}

func ParsePlanetFile(filename string) (*World, error) {
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file (%s) error: %v", filename, err)
	}
	return ParseWorld(fileContent)
}

func ParsePlanetBase64(b64encText string) (*World, error) {
	fileContent, err := base64.StdEncoding.DecodeString(b64encText)
	if err != nil {
		return nil, fmt.Errorf("parse base64 error: %v", err)
	}
	return ParseWorld(fileContent)
}

func demo() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ztplanet <world.bin>")
		os.Exit(1)
	}

	filename := os.Args[1]
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	world, err := ParseWorld(data)
	if err != nil {
		fmt.Printf("Error parsing world file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("ZeroTier Planet Information:\n")
	fmt.Printf("  ID: %d\n", world.ID)
	fmt.Printf("  Type: %d (1=Planet, 127=Moon)\n", world.Type)
	fmt.Printf("  Timestamp: %d\n", world.Timestamp)
	fmt.Printf("  Update Signer Public Key: %s\n", hex.EncodeToString(world.UpdatesMustBeSignedBy[:]))
	fmt.Printf("  Signature: %s...\n", hex.EncodeToString(world.Signature[:16]))
	fmt.Printf("  Number of Roots: %d\n", len(world.Roots))

	for i, root := range world.Roots {
		fmt.Printf("\nRoot Server %d:\n", i+1)
		fmt.Printf("  Identity: %s\n", root.Identity.String())
		for j, ep := range root.StableEndpoints {
			fmt.Printf("  Endpoint %d: %s\n", j+1, ep.String())
		}
	}
}
