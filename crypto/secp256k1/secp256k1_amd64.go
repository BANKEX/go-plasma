package secp256k1

// +build amd64

/*
#cgo CFLAGS: -I./secp256k1
#cgo CFLAGS: -I./secp256k1/src/
#undef USE_ASM_X86_64
#undef USE_ENDOMORPHISM
#if defined(__SIZEOF_INT128__)
#define HAVE___INT128 1
#endif
#define HAVE_BUILTIN_EXPECT 1
#define ENABLE_MODULE_RECOVERY 1
#define USE_NUM_NONE
#define USE_FIELD_5X52
#define USE_FIELD_INV_BUILTIN
#define USE_SCALAR_4X64
#define USE_SCALAR_INV_BUILTIN
#define NDEBUG
#include "./secp256k1/src/secp256k1.c"
#include "./secp256k1/src/modules/recovery/main_impl.h"
#include "bindings.h"
typedef void (*callbackFunc) (const char* msg, void* data);
extern void secp256k1GoPanicIllegal(const char* msg, void* data);
extern void secp256k1GoPanicError(const char* msg, void* data);
*/
import "C"

import (
	"errors"
	"unsafe"
)

type Secp256k1BoundProcessor struct {
	context *C.secp256k1_context
}

func NewSecp256k1BoundProcessor() *Secp256k1BoundProcessor {
	context := C.secp256k1_context_create(C.SECP256K1_CONTEXT_SIGN | C.SECP256K1_CONTEXT_VERIFY)
	C.secp256k1_context_set_illegal_callback(context, C.callbackFunc(C.secp256k1GoPanicIllegal), nil)
	C.secp256k1_context_set_error_callback(context, C.callbackFunc(C.secp256k1GoPanicError), nil)
	new := &Secp256k1BoundProcessor{context: context}
	return new
}

var (
	ErrInvalidMsgLen       = errors.New("invalid message length, need 32 bytes")
	ErrInvalidSignatureLen = errors.New("invalid signature length")
	ErrInvalidRecoveryID   = errors.New("invalid signature recovery id")
	ErrInvalidKey          = errors.New("invalid private key")
	ErrInvalidPubkey       = errors.New("invalid public key")
	ErrSignFailed          = errors.New("signing failed")
	ErrRecoverFailed       = errors.New("recovery failed")
)

// Sign creates a recoverable ECDSA signature.
// The produced signature is in the 65-byte [R || S || V] format where V is 0 or 1.
//
// The caller is responsible for ensuring that msg cannot be chosen
// directly by an attacker. It is usually preferable to use a cryptographic
// hash function on any input before handing it to this function.
func (ctx *Secp256k1BoundProcessor) Sign(msg []byte, seckey []byte) ([]byte, error) {
	context := ctx.context
	if len(msg) != 32 {
		return nil, ErrInvalidMsgLen
	}
	if len(seckey) != 32 {
		return nil, ErrInvalidKey
	}
	seckeydata := (*C.uchar)(unsafe.Pointer(&seckey[0]))
	if C.secp256k1_ec_seckey_verify(context, seckeydata) != 1 {
		return nil, ErrInvalidKey
	}

	var (
		msgdata   = (*C.uchar)(unsafe.Pointer(&msg[0]))
		noncefunc = C.secp256k1_nonce_function_rfc6979
		sigstruct C.secp256k1_ecdsa_recoverable_signature
	)
	if C.secp256k1_ecdsa_sign_recoverable(context, &sigstruct, msgdata, seckeydata, noncefunc, nil) == 0 {
		return nil, ErrSignFailed
	}

	var (
		sig     = make([]byte, 65)
		sigdata = (*C.uchar)(unsafe.Pointer(&sig[0]))
		recid   C.int
	)
	C.secp256k1_ecdsa_recoverable_signature_serialize_compact(context, sigdata, &recid, &sigstruct)
	sig[64] = byte(recid) // add back recid to get 65 bytes sig
	return sig, nil
}

// RecoverPubkey returns the the public key of the signer.
// msg must be the 32-byte hash of the message to be signed.
// sig must be a 65-byte compact ECDSA signature containing the
// recovery id as the last element.
func (ctx *Secp256k1BoundProcessor) RecoverPubkey(msg []byte, sig []byte) ([]byte, error) {
	context := ctx.context
	if len(msg) != 32 {
		return nil, ErrInvalidMsgLen
	}
	if err := checkSignature(sig); err != nil {
		return nil, err
	}

	var (
		pubkey  = make([]byte, 65)
		sigdata = (*C.uchar)(unsafe.Pointer(&sig[0]))
		msgdata = (*C.uchar)(unsafe.Pointer(&msg[0]))
	)
	if C.secp256k1_ext_ecdsa_recover(context, (*C.uchar)(unsafe.Pointer(&pubkey[0])), sigdata, msgdata) == 0 {
		return nil, ErrRecoverFailed
	}
	return pubkey, nil
}

// VerifySignature checks that the given pubkey created signature over message.
// The signature should be in [R || S] format.
func (ctx *Secp256k1BoundProcessor) VerifySignature(pubkey, msg, signature []byte) bool {
	context := ctx.context
	if len(msg) != 32 || len(signature) != 64 || len(pubkey) == 0 {
		return false
	}
	sigdata := (*C.uchar)(unsafe.Pointer(&signature[0]))
	msgdata := (*C.uchar)(unsafe.Pointer(&msg[0]))
	keydata := (*C.uchar)(unsafe.Pointer(&pubkey[0]))
	return C.secp256k1_ext_ecdsa_verify(context, sigdata, msgdata, keydata, C.size_t(len(pubkey))) != 0
}

func checkSignature(sig []byte) error {
	if len(sig) != 65 {
		return ErrInvalidSignatureLen
	}
	if sig[64] >= 4 {
		return ErrInvalidRecoveryID
	}
	return nil
}
