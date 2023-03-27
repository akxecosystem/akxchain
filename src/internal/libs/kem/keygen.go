package kem

import (
	"bytes"
	"errors"
	"github.com/cloudflare/circl/kem"
	"github.com/cloudflare/circl/kem/kyber/kyber768"
	"github.com/cloudflare/circl/kem/schemes"
	"math/rand"
	"time"
)

type KeyGen struct {
	s     kem.Scheme
	seed  []byte
	eseed []byte
	sk    kem.PrivateKey
	pub   kem.PublicKey
}

type Keys struct {
	KeyGen
	Private kem.PrivateKey
	Public  kem.PublicKey
}

type Shared struct {
	CipherText   []byte
	SharedSecret []byte
}

type KemPubKey struct{}

func (kg KeyGen) NewKem() (*Keys, *Shared) {
	kg.s = schemes.ByName("Kyber768")

	kseed := make([]byte, kyber768.KeySeedSize)
	eSeed := make([]byte, kg.s.EncapsulationSeedSize())

	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	r.Read(kseed)
	r.Read(eSeed)
	kg.seed = kseed
	pk, sk := kg.s.DeriveKeyPair(kg.seed)
	kg.sk = sk
	kg.pub = pk

	ct, ss, err := kg.s.EncapsulateDeterministically(kg.pub, eSeed)
	if err != nil {
		panic(err)
	}

	k := &Keys{kg, kg.sk, kg.pub}
	s := &Shared{ct, ss}
	return k, s

}

func (k *Keys) PubBytes() []byte {
	pub, _ := k.Public.MarshalBinary()
	return pub
}

func (k *Keys) PrivBytes() []byte {
	sk, _ := k.Private.MarshalBinary()
	return sk
}

func (k *Keys) GetPublic() kem.PublicKey {
	pub, _ := k.s.UnmarshalBinaryPublicKey(k.PubBytes())
	return pub
}

func (k *Keys) GetPrivate() kem.PrivateKey {
	sk, _ := k.s.UnmarshalBinaryPrivateKey(k.PrivBytes())
	return sk
}

func CreateNewKeys() *Keys {
	var kg KeyGen
	k, _ := kg.NewKem()
	return k
}

func CreateNewKeysWithSharedSecret() (*Keys, *Shared) {
	var kg KeyGen
	k, s := kg.NewKem()
	return k, s
}

func (s *Shared) VerifySharedSecret(k *Keys) error {

	decapsulate, err := k.s.Decapsulate(k.GetPrivate(), s.CipherText)
	if err != nil {
		return err
	}
	if !bytes.Equal(decapsulate, s.SharedSecret) {
		return errors.New("invalid shared secret")
	}
	return nil

}
