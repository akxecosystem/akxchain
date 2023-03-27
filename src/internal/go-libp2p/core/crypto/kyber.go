package crypto

import (
	pb "akxchain/internal/go-libp2p/core/crypto/pb"
	"bytes"
	cryptoRand "crypto/rand"
	kyber "github.com/cloudflare/circl/kem/kyber/kyber1024"
	"github.com/cloudflare/circl/sign/dilithium"
	"github.com/libp2p/go-libp2p/core/crypto"
)

type AkxPrivateKey struct {
	crypto.PrivKey
	priv        []byte
	kPriv       kyber.PrivateKey
	privSignKey []byte
}

type AkxPublicKey struct {
	crypto.PubKey
	pub          []byte
	pubVerifyKey []byte
}

type AkxSignKeyPair struct {
	priv []byte
	pub  []byte
}

type AkxKeyPair struct {
	AkxPrivateKey
	AkxPublicKey
	skp *AkxSignKeyPair
}

func GenerateKyberKeyPair() (*AkxKeyPair, error) {

	pubKey, privKey, err := kyber.GenerateKeyPair(cryptoRand.Reader)
	if err != nil {
		return nil, err
	}
	kp := &AkxKeyPair{}
	var privPack [kyber.PrivateKeySize]byte
	privKey.Pack(privPack[:])
	kp.priv = privPack[:]
	var pubPack [kyber.PublicKeySize]byte
	pubKey.Pack(pubPack[:])
	kp.pub = pubPack[:]
	skp := generateSignKeys()
	kp.privSignKey = skp.priv
	kp.skp = skp
	return kp, nil

}

func (prv *AkxPrivateKey) Raw() []byte {
	return prv.priv
}

func (prv *AkxPrivateKey) Equals(other crypto.Key) bool {
	b, _ := other.Raw()
	return bytes.Equal(prv.priv, b)
}

func (prv *AkxPrivateKey) Type() pb.KeyType {
	return pb.KeyType_KYBER
}

func (prv *AkxPrivateKey) Sign(data []byte) ([]byte, error) {

	sk := getSignKey(prv.privSignKey)
	return sk.Sign(cryptoRand.Reader, data, nil)

}

func (prv *AkxPrivateKey) GetPublic() crypto.PubKey {
	var sk kyber.PrivateKey

	sk.Unpack(prv.priv)
	pub := sk.Public()

	data, err := pub.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return &AkxPublicKey{pub: data}
}

func generateSignKeys() *AkxSignKeyPair {
	pub, priv, err := dilithium.Mode3.GenerateKey(cryptoRand.Reader)
	if err != nil {
		panic(err)
	}
	skp := &AkxSignKeyPair{priv: priv.Bytes(), pub: pub.Bytes()}
	return skp

}

func getSignKey(skpprv []byte) dilithium.PrivateKey {
	return dilithium.Mode3.PrivateKeyFromBytes(skpprv)

}
