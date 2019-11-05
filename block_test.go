// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package qtumsuite_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/qtumproject/qtumsuite"
	"github.com/qtumproject/qtumsuite/chaincfg/chainhash"
	"github.com/qtumproject/qtumsuite/wire"
)

// TestBlock tests the API for Block.
func TestBlock(t *testing.T) {
	b := qtumsuite.NewBlock(&Block99999)

	// Ensure we get the same data back out.
	if msgBlock := b.MsgBlock(); !reflect.DeepEqual(msgBlock, &Block99999) {
		t.Errorf("MsgBlock: mismatched MsgBlock - got %v, want %v",
			spew.Sdump(msgBlock), spew.Sdump(&Block99999))
	}

	// Ensure block height set and get work properly.
	wantHeight := int32(99999)
	b.SetHeight(wantHeight)
	if gotHeight := b.Height(); gotHeight != wantHeight {
		t.Errorf("Height: mismatched height - got %v, want %v",
			gotHeight, wantHeight)
	}

	// Hash for block 100,000.
	wantHashStr := "121d7323ccf8e29595d9a27bb99f6d9a2d41e9301080bb8f0fc221be2e99fa77"
	wantHash, err := chainhash.NewHashFromStr(wantHashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// Request the hash multiple times to test generation and caching.
	for i := 0; i < 2; i++ {
		hash := b.Hash()
		if !hash.IsEqual(wantHash) {
			t.Errorf("Hash #%d mismatched hash - got %v, want %v",
				i, hash, wantHash)
		}
	}

	// Hashes for the transactions in Block99999.
	wantTxHashes := []string{
		"b55b0b1255accaa5cf46a27b06aef05ddd83c55dd98dc212313aaf2d4109ce64",
		"fff2525b8931402dd09222c50775608f75787bd2b87e56995a7bdd30f79702c4",
		"6359f0868171b1d194cbee1af2f16ea598ae8fad666d9b012c8ed2b79a236ec4",
		"e9a66845e05d5abc0ad04ec80f774a7e585c6e8db975962d069a522137b80c1d",
	}

	// Create a new block to nuke all cached data.
	b = qtumsuite.NewBlock(&Block99999)

	// Request hash for all transactions one at a time via Tx.
	for i, txHash := range wantTxHashes {
		wantHash, err := chainhash.NewHashFromStr(txHash)
		if err != nil {
			t.Errorf("NewHashFromStr: %v", err)
		}

		// Request the hash multiple times to test generation and
		// caching.
		for j := 0; j < 2; j++ {
			tx, err := b.Tx(i)
			if err != nil {
				t.Errorf("Tx #%d: %v", i, err)
				continue
			}

			hash := tx.Hash()
			if !hash.IsEqual(wantHash) {
				t.Errorf("Hash #%d mismatched hash - got %v, "+
					"want %v", j, hash, wantHash)
				continue
			}
		}
	}

	// Create a new block to nuke all cached data.
	b = qtumsuite.NewBlock(&Block99999)

	// Request slice of all transactions multiple times to test generation
	// and caching.
	for i := 0; i < 2; i++ {
		transactions := b.Transactions()

		// Ensure we get the expected number of transactions.
		if len(transactions) != len(wantTxHashes) {
			t.Errorf("Transactions #%d mismatched number of "+
				"transactions - got %d, want %d", i,
				len(transactions), len(wantTxHashes))
			continue
		}

		// Ensure all of the hashes match.
		for j, tx := range transactions {
			wantHash, err := chainhash.NewHashFromStr(wantTxHashes[j])
			if err != nil {
				t.Errorf("NewHashFromStr: %v", err)
			}

			hash := tx.Hash()
			if !hash.IsEqual(wantHash) {
				t.Errorf("Transactions #%d mismatched hashes "+
					"- got %v, want %v", j, hash, wantHash)
				continue
			}
		}
	}

	// Serialize the test block.
	var Block99999Buf bytes.Buffer
	err = Block99999.Serialize(&Block99999Buf)
	if err != nil {
		t.Errorf("Serialize: %v", err)
	}
	Block99999Bytes := Block99999Buf.Bytes()

	// Request serialized bytes multiple times to test generation and
	// caching.
	for i := 0; i < 2; i++ {
		serializedBytes, err := b.Bytes()
		if err != nil {
			t.Errorf("Bytes: %v", err)
			continue
		}
		if !bytes.Equal(serializedBytes, Block99999Bytes) {
			t.Errorf("Bytes #%d wrong bytes - got %v, want %v", i,
				spew.Sdump(serializedBytes),
				spew.Sdump(Block99999Bytes))
			continue
		}
	}

	// Transaction offsets and length for the transaction in Block99999.
	wantTxLocs := []wire.TxLoc{
		{TxStart: 253, TxLen: 135},
		{TxStart: 388, TxLen: 259},
		{TxStart: 647, TxLen: 257},
		{TxStart: 904, TxLen: 225},
	}

	// Ensure the transaction location information is accurate.
	txLocs, err := b.TxLoc()
	if err != nil {
		t.Errorf("TxLoc: %v", err)
		return
	}
	if !reflect.DeepEqual(txLocs, wantTxLocs) {
		t.Errorf("TxLoc: mismatched transaction location information "+
			"- got %v, want %v", spew.Sdump(txLocs),
			spew.Sdump(wantTxLocs))
	}
}

// TestNewBlockFromBytes tests creation of a Block from serialized bytes.
func TestNewBlockFromBytes(t *testing.T) {
	// Serialize the test block.
	var Block99999Buf bytes.Buffer
	err := Block99999.Serialize(&Block99999Buf)
	if err != nil {
		t.Errorf("Serialize: %v", err)
	}
	Block99999Bytes := Block99999Buf.Bytes()

	// Create a new block from the serialized bytes.
	b, err := qtumsuite.NewBlockFromBytes(Block99999Bytes)
	if err != nil {
		t.Errorf("NewBlockFromBytes: %v", err)
		return
	}

	// Ensure we get the same data back out.
	serializedBytes, err := b.Bytes()
	if err != nil {
		t.Errorf("Bytes: %v", err)
		return
	}
	if !bytes.Equal(serializedBytes, Block99999Bytes) {
		t.Errorf("Bytes: wrong bytes - got %v, want %v",
			spew.Sdump(serializedBytes),
			spew.Sdump(Block99999Bytes))
	}

	// Ensure the generated MsgBlock is correct.
	if msgBlock := b.MsgBlock(); !reflect.DeepEqual(msgBlock, &Block99999) {
		t.Errorf("MsgBlock: mismatched MsgBlock - got %v, want %v",
			spew.Sdump(msgBlock), spew.Sdump(&Block99999))
	}
}

// TestNewBlockFromBlockAndBytes tests creation of a Block from a MsgBlock and
// raw bytes.
func TestNewBlockFromBlockAndBytes(t *testing.T) {
	// Serialize the test block.
	var Block99999Buf bytes.Buffer
	err := Block99999.Serialize(&Block99999Buf)
	if err != nil {
		t.Errorf("Serialize: %v", err)
	}
	Block99999Bytes := Block99999Buf.Bytes()

	// Create a new block from the serialized bytes.
	b := qtumsuite.NewBlockFromBlockAndBytes(&Block99999, Block99999Bytes)

	// Ensure we get the same data back out.
	serializedBytes, err := b.Bytes()
	if err != nil {
		t.Errorf("Bytes: %v", err)
		return
	}
	if !bytes.Equal(serializedBytes, Block99999Bytes) {
		t.Errorf("Bytes: wrong bytes - got %v, want %v",
			spew.Sdump(serializedBytes),
			spew.Sdump(Block99999Bytes))
	}
	if msgBlock := b.MsgBlock(); !reflect.DeepEqual(msgBlock, &Block99999) {
		t.Errorf("MsgBlock: mismatched MsgBlock - got %v, want %v",
			spew.Sdump(msgBlock), spew.Sdump(&Block99999))
	}
}

// TestBlockErrors tests the error paths for the Block API.
func TestBlockErrors(t *testing.T) {
	// Ensure out of range errors are as expected.
	wantErr := "transaction index -1 is out of range - max 3"
	testErr := qtumsuite.OutOfRangeError(wantErr)
	if testErr.Error() != wantErr {
		t.Errorf("OutOfRangeError: wrong error - got %v, want %v",
			testErr.Error(), wantErr)
	}

	// Serialize the test block.
	var Block99999Buf bytes.Buffer
	err := Block99999.Serialize(&Block99999Buf)
	if err != nil {
		t.Errorf("Serialize: %v", err)
	}
	Block99999Bytes := Block99999Buf.Bytes()

	// Create a new block from the serialized bytes.
	b, err := qtumsuite.NewBlockFromBytes(Block99999Bytes)
	if err != nil {
		t.Errorf("NewBlockFromBytes: %v", err)
		return
	}

	// Truncate the block byte buffer to force errors.
	shortBytes := Block99999Bytes[:80]
	_, err = qtumsuite.NewBlockFromBytes(shortBytes)
	if err != io.EOF {
		t.Errorf("NewBlockFromBytes: did not get expected error - "+
			"got %v, want %v", err, io.EOF)
	}

	// Ensure TxHash returns expected error on invalid indices.
	_, err = b.TxHash(-1)
	if _, ok := err.(qtumsuite.OutOfRangeError); !ok {
		t.Errorf("TxHash: wrong error - got: %v <%T>, "+
			"want: <%T>", err, err, qtumsuite.OutOfRangeError(""))
	}
	_, err = b.TxHash(len(Block99999.Transactions) + 1)
	if _, ok := err.(qtumsuite.OutOfRangeError); !ok {
		t.Errorf("TxHash: wrong error - got: %v <%T>, "+
			"want: <%T>", err, err, qtumsuite.OutOfRangeError(""))
	}

	// Ensure Tx returns expected error on invalid indices.
	_, err = b.Tx(-1)
	if _, ok := err.(qtumsuite.OutOfRangeError); !ok {
		t.Errorf("Tx: wrong error - got: %v <%T>, "+
			"want: <%T>", err, err, qtumsuite.OutOfRangeError(""))
	}
	_, err = b.Tx(len(Block99999.Transactions) + 1)
	if _, ok := err.(qtumsuite.OutOfRangeError); !ok {
		t.Errorf("Tx: wrong error - got: %v <%T>, "+
			"want: <%T>", err, err, qtumsuite.OutOfRangeError(""))
	}

	// Ensure TxLoc returns expected error with short byte buffer.
	// This makes use of the test package only function, SetBlockBytes, to
	// inject a short byte buffer.
	b.SetBlockBytes(shortBytes)
	_, err = b.TxLoc()
	if err != io.EOF {
		t.Errorf("TxLoc: did not get expected error - "+
			"got %v, want %v", err, io.EOF)
	}
}

// Block99999 defines block 99999 of the block chain.  It is used to
// test Block operations.
var Block99999 = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version: 2,
		PrevBlock: chainhash.Hash([32]byte{ // Make go vet happy.
			0x2b, 0x24, 0xbd, 0xd0, 0xa7, 0x88, 0xeb, 0x1e,
			0xdb, 0x44, 0x25, 0xe6, 0xc1, 0x94, 0x1a, 0x58,
			0x62, 0xd8, 0x27, 0xa2, 0x05, 0x87, 0x49, 0xcc,
			0x86, 0x08, 0x2d, 0x58, 0x27, 0x64, 0xdf, 0x2d,
		}), // 2ddf6427582d0886cc498705a227d862581a94c1e62544db1eeb88a7d0bd242b
		MerkleRoot: chainhash.Hash([32]byte{ // Make go vet happy.
			0xc5, 0xee, 0x42, 0x1c, 0x2b, 0x6e, 0x43, 0x8c,
			0x4a, 0x1f, 0x3d, 0x1b, 0xe1, 0x15, 0xb8, 0xb1,
			0x43, 0x72, 0x07, 0xc4, 0xb6, 0x86, 0x52, 0xcf,
			0x22, 0xff, 0x8d, 0x7a, 0x98, 0x93, 0x8f, 0xb0,
		}), // b08f93987a8dff22cf5286b6c4077243b1b815e11b3d1f4a8c436e2b1c42eec5
		Timestamp: time.Unix(1518578512, 0), // 2010-12-29 11:57:43 +0000 UTC
		Bits:      0x1a053178,               // 453281356
		Nonce:     0,                        // 274148111
		HashStateRoot: chainhash.Hash([32]byte{
			0xb5, 0xd0, 0x3b, 0xdd, 0xfc, 0x31, 0x97, 0xa1,
			0xf1, 0xac, 0xc5, 0xb7, 0x5a, 0xe3, 0x42, 0x25,
			0x39, 0x03, 0xeb, 0xde, 0x0f, 0x5d, 0x96, 0xf8,
			0x81, 0xb8, 0x6b, 0x8b, 0x1f, 0x43, 0x1b, 0xf5,
		}),
		HashUTXORoot: chainhash.Hash([32]byte{
			0xf5, 0xf2, 0x2c, 0x17, 0xb5, 0x37, 0xd9, 0x84,
			0xf5, 0x91, 0xae, 0x92, 0x6a, 0x35, 0xf2, 0x96,
			0x9f, 0x8d, 0x98, 0xa9, 0xc9, 0xd3, 0x55, 0x9b,
			0x9d, 0x1c, 0x1a, 0x87, 0x65, 0xba, 0xb4, 0x61,
		}),
		PrevoutStake: wire.OutPoint{
			Hash: chainhash.Hash([32]byte{
				0xee, 0x46, 0xfb, 0x2c, 0x46, 0x2c, 0xa1, 0x78,
				0x05, 0xe6, 0x8c, 0xbb, 0x96, 0xa1, 0x64, 0xd0,
				0xee, 0xb5, 0x30, 0x2d, 0x93, 0x50, 0x59, 0xdf,
				0x71, 0x0e, 0xf1, 0xa7, 0xb3, 0x77, 0x20, 0xd1,
			}),
			Index: 2,
		},
		BlockSig: []byte{
			0x30, 0x45, 0x02, 0x21, 0x00, 0xbe, 0xac, 0x3b,
			0x08, 0x1a, 0x8b, 0xe6, 0x17, 0x82, 0x74, 0x12,
			0xc1, 0xb8, 0x76, 0xdf, 0xab, 0x2a, 0x1b, 0x64,
			0x80, 0x1f, 0xfb, 0x5d, 0xa1, 0xaf, 0xdc, 0x94,
			0x05, 0x2b, 0xf0, 0x75, 0x37, 0x02, 0x20, 0x37,
			0x18, 0xe6, 0xf6, 0xed, 0x44, 0x76, 0xbc, 0xa2,
			0x8e, 0x7a, 0xa9, 0x9c, 0x10, 0x78, 0x0f, 0x22,
			0x39, 0x51, 0xf1, 0xf9, 0x1f, 0xe0, 0x55, 0x43,
			0x77, 0x17, 0x0c, 0x19, 0x23, 0x79, 0xba,
		},
	},
	Transactions: []*wire.MsgTx{
		{
			Version: 2,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash:  chainhash.Hash{},
						Index: 0xffffffff,
					},
					SignatureScript: []byte{
						0x04, 0x4c, 0x86, 0x04, 0x1b, 0x02, 0x06, 0x02,
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0x12a05f200, // 5000000000
					PkScript: []byte{
						0x41, // OP_DATA_65
						0x04, 0x1b, 0x0e, 0x8c, 0x25, 0x67, 0xc1, 0x25,
						0x36, 0xaa, 0x13, 0x35, 0x7b, 0x79, 0xa0, 0x73,
						0xdc, 0x44, 0x44, 0xac, 0xb8, 0x3c, 0x4e, 0xc7,
						0xa0, 0xe2, 0xf9, 0x9d, 0xd7, 0x45, 0x75, 0x16,
						0xc5, 0x81, 0x72, 0x42, 0xda, 0x79, 0x69, 0x24,
						0xca, 0x4e, 0x99, 0x94, 0x7d, 0x08, 0x7f, 0xed,
						0xf9, 0xce, 0x46, 0x7c, 0xb9, 0xf7, 0xc6, 0x28,
						0x70, 0x78, 0xf8, 0x01, 0xdf, 0x27, 0x6f, 0xdf,
						0x84, // 65-byte signature
						0xac, // OP_CHECKSIG
					},
				},
			},
			LockTime: 0,
		},
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash: chainhash.Hash([32]byte{ // Make go vet happy.
							0x03, 0x2e, 0x38, 0xe9, 0xc0, 0xa8, 0x4c, 0x60,
							0x46, 0xd6, 0x87, 0xd1, 0x05, 0x56, 0xdc, 0xac,
							0xc4, 0x1d, 0x27, 0x5e, 0xc5, 0x5f, 0xc0, 0x07,
							0x79, 0xac, 0x88, 0xfd, 0xf3, 0x57, 0xa1, 0x87,
						}), // 87a157f3fd88ac7907c05fc55e271dc4acdc5605d187d646604ca8c0e9382e03
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xc3, 0x52, 0xd3,
						0xdd, 0x99, 0x3a, 0x98, 0x1b, 0xeb, 0xa4, 0xa6,
						0x3a, 0xd1, 0x5c, 0x20, 0x92, 0x75, 0xca, 0x94,
						0x70, 0xab, 0xfc, 0xd5, 0x7d, 0xa9, 0x3b, 0x58,
						0xe4, 0xeb, 0x5d, 0xce, 0x82, 0x02, 0x21, 0x00,
						0x84, 0x07, 0x92, 0xbc, 0x1f, 0x45, 0x60, 0x62,
						0x81, 0x9f, 0x15, 0xd3, 0x3e, 0xe7, 0x05, 0x5c,
						0xf7, 0xb5, 0xee, 0x1a, 0xf1, 0xeb, 0xcc, 0x60,
						0x28, 0xd9, 0xcd, 0xb1, 0xc3, 0xaf, 0x77, 0x48,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xf4, 0x6d, 0xb5, 0xe9, 0xd6, 0x1a, 0x9d,
						0xc2, 0x7b, 0x8d, 0x64, 0xad, 0x23, 0xe7, 0x38,
						0x3a, 0x4e, 0x6c, 0xa1, 0x64, 0x59, 0x3c, 0x25,
						0x27, 0xc0, 0x38, 0xc0, 0x85, 0x7e, 0xb6, 0x7e,
						0xe8, 0xe8, 0x25, 0xdc, 0xa6, 0x50, 0x46, 0xb8,
						0x2c, 0x93, 0x31, 0x58, 0x6c, 0x82, 0xe0, 0xfd,
						0x1f, 0x63, 0x3f, 0x25, 0xf8, 0x7c, 0x16, 0x1b,
						0xc6, 0xf8, 0xa6, 0x30, 0x12, 0x1d, 0xf2, 0xb3,
						0xd3, // 65-byte pubkey
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0x2123e300, // 556000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xc3, 0x98, 0xef, 0xa9, 0xc3, 0x92, 0xba, 0x60,
						0x13, 0xc5, 0xe0, 0x4e, 0xe7, 0x29, 0x75, 0x5e,
						0xf7, 0xf5, 0x8b, 0x32,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
				{
					Value: 0x108e20f00, // 4444000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x94, 0x8c, 0x76, 0x5a, 0x69, 0x14, 0xd4, 0x3f,
						0x2a, 0x7a, 0xc1, 0x77, 0xda, 0x2c, 0x2f, 0x6b,
						0x52, 0xde, 0x3d, 0x7c,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
			},
			LockTime: 0,
		},
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash: chainhash.Hash([32]byte{ // Make go vet happy.
							0xc3, 0x3e, 0xbf, 0xf2, 0xa7, 0x09, 0xf1, 0x3d,
							0x9f, 0x9a, 0x75, 0x69, 0xab, 0x16, 0xa3, 0x27,
							0x86, 0xaf, 0x7d, 0x7e, 0x2d, 0xe0, 0x92, 0x65,
							0xe4, 0x1c, 0x61, 0xd0, 0x78, 0x29, 0x4e, 0xcf,
						}), // cf4e2978d0611ce46592e02d7e7daf8627a316ab69759a9f3df109a7f2bf3ec3
						Index: 1,
					},
					SignatureScript: []byte{
						0x47, // OP_DATA_71
						0x30, 0x44, 0x02, 0x20, 0x03, 0x2d, 0x30, 0xdf,
						0x5e, 0xe6, 0xf5, 0x7f, 0xa4, 0x6c, 0xdd, 0xb5,
						0xeb, 0x8d, 0x0d, 0x9f, 0xe8, 0xde, 0x6b, 0x34,
						0x2d, 0x27, 0x94, 0x2a, 0xe9, 0x0a, 0x32, 0x31,
						0xe0, 0xba, 0x33, 0x3e, 0x02, 0x20, 0x3d, 0xee,
						0xe8, 0x06, 0x0f, 0xdc, 0x70, 0x23, 0x0a, 0x7f,
						0x5b, 0x4a, 0xd7, 0xd7, 0xbc, 0x3e, 0x62, 0x8c,
						0xbe, 0x21, 0x9a, 0x88, 0x6b, 0x84, 0x26, 0x9e,
						0xae, 0xb8, 0x1e, 0x26, 0xb4, 0xfe, 0x01,
						0x41, // OP_DATA_65
						0x04, 0xae, 0x31, 0xc3, 0x1b, 0xf9, 0x12, 0x78,
						0xd9, 0x9b, 0x83, 0x77, 0xa3, 0x5b, 0xbc, 0xe5,
						0xb2, 0x7d, 0x9f, 0xff, 0x15, 0x45, 0x68, 0x39,
						0xe9, 0x19, 0x45, 0x3f, 0xc7, 0xb3, 0xf7, 0x21,
						0xf0, 0xba, 0x40, 0x3f, 0xf9, 0x6c, 0x9d, 0xee,
						0xb6, 0x80, 0xe5, 0xfd, 0x34, 0x1c, 0x0f, 0xc3,
						0xa7, 0xb9, 0x0d, 0xa4, 0x63, 0x1e, 0xe3, 0x95,
						0x60, 0x63, 0x9d, 0xb4, 0x62, 0xe9, 0xcb, 0x85,
						0x0f, // 65-byte pubkey
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0xf4240, // 1000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xb0, 0xdc, 0xbf, 0x97, 0xea, 0xbf, 0x44, 0x04,
						0xe3, 0x1d, 0x95, 0x24, 0x77, 0xce, 0x82, 0x2d,
						0xad, 0xbe, 0x7e, 0x10,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
				{
					Value: 0x11d260c0, // 299000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x6b, 0x12, 0x81, 0xee, 0xc2, 0x5a, 0xb4, 0xe1,
						0xe0, 0x79, 0x3f, 0xf4, 0xe0, 0x8a, 0xb1, 0xab,
						0xb3, 0x40, 0x9c, 0xd9,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
			},
			LockTime: 0,
		},
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash: chainhash.Hash([32]byte{ // Make go vet happy.
							0x0b, 0x60, 0x72, 0xb3, 0x86, 0xd4, 0xa7, 0x73,
							0x23, 0x52, 0x37, 0xf6, 0x4c, 0x11, 0x26, 0xac,
							0x3b, 0x24, 0x0c, 0x84, 0xb9, 0x17, 0xa3, 0x90,
							0x9b, 0xa1, 0xc4, 0x3d, 0xed, 0x5f, 0x51, 0xf4,
						}), // f4515fed3dc4a19b90a317b9840c243bac26114cf637522373a7d486b372600b
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xbb, 0x1a, 0xd2,
						0x6d, 0xf9, 0x30, 0xa5, 0x1c, 0xce, 0x11, 0x0c,
						0xf4, 0x4f, 0x7a, 0x48, 0xc3, 0xc5, 0x61, 0xfd,
						0x97, 0x75, 0x00, 0xb1, 0xae, 0x5d, 0x6b, 0x6f,
						0xd1, 0x3d, 0x0b, 0x3f, 0x4a, 0x02, 0x21, 0x00,
						0xc5, 0xb4, 0x29, 0x51, 0xac, 0xed, 0xff, 0x14,
						0xab, 0xba, 0x27, 0x36, 0xfd, 0x57, 0x4b, 0xdb,
						0x46, 0x5f, 0x3e, 0x6f, 0x8d, 0xa1, 0x2e, 0x2c,
						0x53, 0x03, 0x95, 0x4a, 0xca, 0x7f, 0x78, 0xf3,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xa7, 0x13, 0x5b, 0xfe, 0x82, 0x4c, 0x97,
						0xec, 0xc0, 0x1e, 0xc7, 0xd7, 0xe3, 0x36, 0x18,
						0x5c, 0x81, 0xe2, 0xaa, 0x2c, 0x41, 0xab, 0x17,
						0x54, 0x07, 0xc0, 0x94, 0x84, 0xce, 0x96, 0x94,
						0xb4, 0x49, 0x53, 0xfc, 0xb7, 0x51, 0x20, 0x65,
						0x64, 0xa9, 0xc2, 0x4d, 0xd0, 0x94, 0xd4, 0x2f,
						0xdb, 0xfd, 0xd5, 0xaa, 0xd3, 0xe0, 0x63, 0xce,
						0x6a, 0xf4, 0xcf, 0xaa, 0xea, 0x4e, 0xa1, 0x4f,
						0xbb, // 65-byte pubkey
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0xf4240, // 1000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x39, 0xaa, 0x3d, 0x56, 0x9e, 0x06, 0xa1, 0xd7,
						0x92, 0x6d, 0xc4, 0xbe, 0x11, 0x93, 0xc9, 0x9b,
						0xf2, 0xeb, 0x9e, 0xe0,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
			},
			LockTime: 0,
		},
	},
}
