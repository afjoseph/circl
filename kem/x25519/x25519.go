package x25519

import (
	"github.com/cloudflare/circl/dh/x25519"
	"github.com/cloudflare/circl/kem"

	"bytes"
	cryptoRand "crypto/rand"
	"crypto/subtle"
)

type PrivateKey x25519.Key

type PublicKey x25519.Key

type scheme struct{}

var Scheme kem.Scheme = &scheme{}

func (*scheme) Name() string               { return "X25519" }
func (*scheme) PublicKeySize() int         { return x25519.Size }
func (*scheme) PrivateKeySize() int        { return x25519.Size }
func (*scheme) SeedSize() int              { return x25519.Size }
func (*scheme) SharedKeySize() int         { return x25519.Size }
func (*scheme) CiphertextSize() int        { return x25519.Size }
func (*scheme) EncapsulationSeedSize() int { return x25519.Size }

func (sk *PrivateKey) Scheme() kem.Scheme { return Scheme }
func (pk *PublicKey) Scheme() kem.Scheme  { return Scheme }

func (sk *PrivateKey) MarshalBinary() ([]byte, error) {
	var ret [x25519.Size]byte
	copy(ret[:], sk[:])
	return ret[:], nil
}

func (sk *PrivateKey) Equal(other kem.PrivateKey) bool {
	oth, ok := other.(*PrivateKey)
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare(sk[:], oth[:]) == 1
}

func (pk *PublicKey) Equal(other kem.PublicKey) bool {
	oth, ok := other.(*PublicKey)
	if !ok {
		return false
	}
	return bytes.Equal(pk[:], oth[:])
}

func (pk *PublicKey) MarshalBinary() ([]byte, error) {
	var ret [x25519.Size]byte
	copy(ret[:], pk[:])
	return ret[:], nil
}

func (sch *scheme) GenerateKey() (kem.PublicKey, kem.PrivateKey, error) {
	var seed [x25519.Size]byte
	_, err := cryptoRand.Read(seed[:])
	if err != nil {
		return nil, nil, err
	}
	pk, sk := sch.DeriveKey(seed[:])
	return pk, sk, nil
}

func (*scheme) DeriveKey(seed []byte) (kem.PublicKey, kem.PrivateKey) {
	if len(seed) != x25519.Size {
		panic(kem.ErrSeedSize)
	}
	var pk PublicKey
	var sk PrivateKey
	copy(sk[:], seed[:])
	x25519.KeyGen((*x25519.Key)(&pk), (*x25519.Key)(&sk))
	return &pk, &sk
}

func (sch *scheme) Encapsulate(pk kem.PublicKey) (ct, ss []byte, err error) {
	var seed [x25519.Size]byte
	_, err = cryptoRand.Read(seed[:])
	if err != nil {
		return
	}
	return sch.EncapsulateDeterministically(pk, seed[:])
}

func (*scheme) EncapsulateDeterministically(pk kem.PublicKey, seed []byte) (
	[]byte, []byte, error) {
	if len(seed) != x25519.Size {
		return nil, nil, kem.ErrSeedSize
	}

	var ct, ss, priv x25519.Key

	copy(priv[:], seed)

	pub, ok := pk.(*PublicKey)
	if !ok {
		return nil, nil, kem.ErrTypeMismatch
	}

	x25519.KeyGen(&ct, &priv)
	if !x25519.Shared(&ss, &priv, (*x25519.Key)(pub)) {
		return nil, nil, kem.ErrPubKey
	}

	return ct[:], ss[:], nil
}

func (*scheme) Decapsulate(sk kem.PrivateKey, ct []byte) ([]byte, error) {
	if len(ct) != x25519.Size {
		return nil, kem.ErrCiphertextSize
	}

	priv, ok := sk.(*PrivateKey)
	if !ok {
		return nil, kem.ErrTypeMismatch
	}

	var ss, ct2 x25519.Key
	copy(ct2[:], ct)

	if !x25519.Shared(&ss, (*x25519.Key)(priv), &ct2) {
		return nil, kem.ErrCipherText
	}

	return ss[:], nil
}

func (*scheme) UnmarshalBinaryPublicKey(buf []byte) (kem.PublicKey, error) {
	if len(buf) != x25519.Size {
		return nil, kem.ErrPubKeySize
	}
	var ret PublicKey
	copy(ret[:], buf)
	return &ret, nil
}

func (*scheme) UnmarshalBinaryPrivateKey(buf []byte) (kem.PrivateKey, error) {
	if len(buf) != x25519.Size {
		return nil, kem.ErrPrivKeySize
	}
	var ret PrivateKey
	copy(ret[:], buf)
	return &ret, nil
}
