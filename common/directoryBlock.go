// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/FactomProject/factoid/block"
	"reflect"
	"sync"
)

const DBlockVersion = 0

type DChain struct {
	ChainID      *Hash
	Blocks       []*DirectoryBlock
	BlockMutex   sync.Mutex
	NextBlock    *DirectoryBlock
	NextDBHeight uint32
	IsValidated  bool
}

func NewDChain() *DChain {
	d := new(DChain)
	d.Blocks = make([]*DirectoryBlock, 0)
	d.NextBlock = NewDirectoryBlock()

	return d
}

type DirectoryBlock struct {
	//Marshalized
	Header    *DBlockHeader
	DBEntries []*DBEntry

	//Not Marshalized
	Chain       *DChain
	IsSealed    bool
	DBHash      *Hash
	KeyMR       *Hash
	IsSavedInDB bool
	IsValidated bool
}

func NewDirectoryBlock() *DirectoryBlock {
	d := new(DirectoryBlock)
	d.Header = NewDBlockHeader()

	d.DBEntries = make([]*DBEntry, 0)
	d.DBHash = NewHash()
	d.KeyMR = NewHash()

	return d
}

func NewDBlock() *DirectoryBlock {
	return NewDirectoryBlock()
}

type DirBlockInfo struct {

	// Serial hash for the directory block
	DBHash *Hash

	// BTCTxHash is the Tx hash returned from rpcclient.SendRawTransaction
	BTCTxHash *Hash // use string or *btcwire.ShaHash ???

	// BTCTxOffset is the index of the TX in this BTC block
	BTCTxOffset int

	// BTCBlockHeight is the height of the block where this TX is stored in BTC
	BTCBlockHeight int32

	//BTCBlockHash is the hash of the block where this TX is stored in BTC
	//BTCBlockHash *Hash // use string or *btcwire.ShaHash ???

	// DBMerkleRoot is the merkle root of the Directory Block
	// and is written into BTC as OP_RETURN data
	DBMerkleRoot *Hash

	// A flag to to show BTC anchor confirmation
	BTCConfirmed bool
}

type DBlockHeader struct {
	Version   byte
	NetworkID uint32

	BodyMR       *Hash
	PrevKeyMR    *Hash
	PrevFullHash *Hash

	Timestamp  uint32
	DBHeight   uint32
	BlockCount uint32
}

func NewDBlockHeader() *DBlockHeader {
	d := new(DBlockHeader)
	d.BodyMR = NewHash()
	d.PrevKeyMR = NewHash()
	d.PrevFullHash = NewHash()

	return d
}

type DBEntry struct {
	ChainID *Hash
	KeyMR   *Hash // Different MR in EBlockHeader
}

func NewDBEntry(eb *EBlock) *DBEntry {
	e := new(DBEntry)

	e.ChainID = eb.Header.ChainID
	e.KeyMR = eb.KeyMR()

	return e
}

func NewDBEntryFromECBlock(cb *ECBlock) *DBEntry {
	e := &DBEntry{}

	e.ChainID = cb.Header.ECChainID
	e.KeyMR = cb.Header.Hash()

	return e
}

func NewDBEntryFromABlock(b *AdminBlock) *DBEntry {
	e := &DBEntry{}

	e.ChainID = b.Header.AdminChainID
	e.KeyMR = b.ABHash

	return e
}

func NewDirBlockInfoFromDBlock(b *DirectoryBlock) *DirBlockInfo {
	e := &DirBlockInfo{}
	e.DBHash = b.DBHash
	e.DBMerkleRoot = b.KeyMR
	e.BTCConfirmed = false
	e.BTCTxHash = NewHash()

	return e
}

//func (e *DBEntry) Hash() *Hash {
//	return e.hash
//}
//
//func (e *DBEntry) SetHash(binaryHash []byte) {
//	h := new(Hash)
//	h.SetBytes(binaryHash)
//	e.hash = h
//}
//
//func (e *DBEntry) EncodableFields() map[string]reflect.Value {
//	fields := map[string]reflect.Value{
//		`KeyMR`:   reflect.ValueOf(e.KeyMR),
//		`ChainID`: reflect.ValueOf(e.ChainID),
//	}
//	return fields
//}

func (e *DBEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, err = e.ChainID.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	data, err = e.KeyMR.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	return buf.Bytes(), nil
}

func (e *DBEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	e.ChainID = new(Hash)
	newData, err = e.ChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	e.KeyMR = new(Hash)
	newData, err = e.KeyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	return
}

func (e *DBEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *DBEntry) ShaHash() *Hash {
	byteArray, _ := e.MarshalBinary()
	return Sha(byteArray)
}

func (b *DBlockHeader) EncodableFields() map[string]reflect.Value {
	fields := map[string]reflect.Value{
		`DBHeight`:     reflect.ValueOf(b.DBHeight),
		`BlockCount`:   reflect.ValueOf(b.BlockCount),
		`BodyMR`:       reflect.ValueOf(b.BodyMR),
		`PrevFullHash`: reflect.ValueOf(b.PrevFullHash),
	}
	return fields
}

func (b *DBlockHeader) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{b.Version})
	binary.Write(&buf, binary.BigEndian, b.NetworkID)

	if b.BodyMR == nil {
		b.BodyMR = new(Hash)
		b.BodyMR.SetBytes(new([32]byte)[:])
	}
	data, err = b.BodyMR.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	data, err = b.PrevKeyMR.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	data, err = b.PrevFullHash.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, b.Timestamp)

	binary.Write(&buf, binary.BigEndian, b.DBHeight)

	binary.Write(&buf, binary.BigEndian, b.BlockCount)

	return buf.Bytes(), err
}

func (b *DBlockHeader) MarshalledSize() uint64 {
	var size uint64 = 0
	size += 1 //Version
	size += 4 //NetworkID
	size += uint64(HASH_LENGTH)
	size += uint64(HASH_LENGTH)
	size += uint64(HASH_LENGTH)
	size += 4 //Timestamp
	size += 4 //DBHeight
	size += 4 //BlockCount

	return size
}

func (b *DBlockHeader) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	b.Version, newData = newData[0], newData[1:]

	b.NetworkID, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	b.BodyMR = new(Hash)
	newData, err = b.BodyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.PrevKeyMR = new(Hash)
	newData, err = b.PrevKeyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.PrevFullHash = new(Hash)
	newData, err = b.PrevFullHash.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.Timestamp, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	b.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	b.BlockCount, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	return
}

func (b *DBlockHeader) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

func CreateDBlock(chain *DChain, prev *DirectoryBlock, cap uint) (b *DirectoryBlock, err error) {
	if prev == nil && chain.NextDBHeight != 0 {
		return nil, errors.New("Previous block cannot be nil")
	} else if prev != nil && chain.NextDBHeight == 0 {
		return nil, errors.New("Origin block cannot have a parent block")
	}

	b = new(DirectoryBlock)

	b.Header = new(DBlockHeader)
	b.Header.Version = VERSION_0

	if prev == nil {
		b.Header.PrevFullHash = NewHash()
		b.Header.PrevKeyMR = NewHash()
	} else {
		b.Header.PrevFullHash, err = CreateHash(prev)
		if prev.KeyMR == nil {
			prev.BuildKeyMerkleRoot()
		}
		b.Header.PrevKeyMR = prev.KeyMR
	}

	b.Header.DBHeight = chain.NextDBHeight
	b.Chain = chain
	b.DBEntries = make([]*DBEntry, 0, cap)
	b.IsSealed = false

	return b, err
}

// Add DBEntry from an Entry Block
func (c *DChain) AddEBlockToDBEntry(eb *EBlock) (err error) {

	dbEntry := NewDBEntry(eb)
	c.BlockMutex.Lock()
	c.NextBlock.DBEntries = append(c.NextBlock.DBEntries, dbEntry)
	c.BlockMutex.Unlock()

	return nil
}

// Add DBEntry from an Entry Credit Block
func (c *DChain) AddECBlockToDBEntry(ecb *ECBlock) (err error) {

	dbEntry := NewDBEntryFromECBlock(ecb)

	if len(c.NextBlock.DBEntries) < 3 {
		panic("1 DBEntries not initialized properly for block: " + string(c.NextDBHeight))
	}

	c.BlockMutex.Lock()
	// Cblock is always at the first entry
	c.NextBlock.DBEntries[1] = dbEntry // First three entries are ABlock, CBlock, FBlock
	c.BlockMutex.Unlock()

	return nil
}

// Add DBEntry from an Admin Block
func (c *DChain) AddABlockToDBEntry(b *AdminBlock) (err error) {

	dbEntry := &DBEntry{}
	dbEntry.ChainID = b.Header.AdminChainID
	dbEntry.KeyMR = b.ABHash

	if len(c.NextBlock.DBEntries) < 3 {
		panic("2 DBEntries not initialized properly for block: " + string(c.NextDBHeight))
	}

	c.BlockMutex.Lock()
	// Ablock is always at the first entry
	// First three entries are ABlock, CBlock, FBlock
	c.NextBlock.DBEntries[0] = dbEntry
	c.BlockMutex.Unlock()

	return nil
}

// Add DBEntry from an SC Block
func (c *DChain) AddFBlockToDBEntry(b block.IFBlock) (err error) {

	dbEntry := &DBEntry{}
	dbEntry.ChainID = new(Hash)
	dbEntry.ChainID.SetBytes(b.GetChainID().Bytes())

	dbEntry.KeyMR = new(Hash)
	dbEntry.KeyMR.SetBytes(b.GetHash().Bytes())

	if len(c.NextBlock.DBEntries) < 3 {
		panic("3 DBEntries not initialized properly for block: " + string(c.NextDBHeight))
	}

	c.BlockMutex.Lock()
	// Ablock is always at the first entry
	// First three entries are ABlock, CBlock, FBlock
	c.NextBlock.DBEntries[2] = dbEntry
	c.BlockMutex.Unlock()

	return nil
}

// Add DBEntry
func (c *DChain) AddDBEntry(dbEntry *DBEntry) (err error) {

	c.BlockMutex.Lock()
	c.NextBlock.DBEntries = append(c.NextBlock.DBEntries, dbEntry)
	c.BlockMutex.Unlock()

	return nil
}

/*
// Add DBEntry from a Factoid Block
func (c *DChain) AddFBlockMRToDBEntry(dbEntry *DBEntry) (err error) {

	fmt.Println("AddFDBlock >>>>>")

	if len(c.NextBlock.DBEntries) < 3 {
		panic("4 DBEntries not initialized properly for block: " + string(c.NextDBHeight))
	}
	c.BlockMutex.Lock()
	// Factoid entry is alwasy at the same position
	// First three entries are ABlock, CBlock, FBlock
	//c.NextBlock.DBEntries[2] = dbEntry
	c.BlockMutex.Unlock()

	return nil
}
*/
// Add DBlock to the chain in memory
func (c *DChain) AddDBlockToDChain(b *DirectoryBlock) (err error) {

	// Increase the slice capacity if needed
	if b.Header.DBHeight >= uint32(cap(c.Blocks)) {
		temp := make([]*DirectoryBlock, len(c.Blocks), b.Header.DBHeight*2)
		copy(temp, c.Blocks)
		c.Blocks = temp
	}

	// Increase the slice length if needed
	if b.Header.DBHeight >= uint32(len(c.Blocks)) {
		c.Blocks = c.Blocks[0 : b.Header.DBHeight+1]
	}

	c.Blocks[b.Header.DBHeight] = b

	return nil
}

// Check if the block with the input block height is existing in chain
func (c *DChain) IsBlockExisting(height uint32) bool {

	if height >= uint32(len(c.Blocks)) {
		return false
	} else if c.Blocks[height] != nil {
		return true
	}

	return false
}

func (b *DirectoryBlock) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, err = b.Header.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	count := uint32(len(b.DBEntries))
	for i := uint32(0); i < count; i = i + 1 {
		data, err = b.DBEntries[i].MarshalBinary()
		if err != nil {
			return
		}
		buf.Write(data)
	}

	return buf.Bytes(), err
}

func (b *DirectoryBlock) BuildBodyMR() (mr *Hash, err error) {
	hashes := make([]*Hash, len(b.DBEntries))
	for i, entry := range b.DBEntries {
		data, _ := entry.MarshalBinary()
		hashes[i] = Sha(data)
	}

	if len(hashes) == 0 {
		hashes = append(hashes, Sha(nil))
	}

	merkle := BuildMerkleTreeStore(hashes)
	return merkle[len(merkle)-1], nil
}

func (b *DirectoryBlock) BuildKeyMerkleRoot() (err error) {

	// Create the Entry Block Key Merkle Root from the hash of Header and the Body Merkle Root
	hashes := make([]*Hash, 0, 2)
	binaryEBHeader, _ := b.Header.MarshalBinary()
	hashes = append(hashes, Sha(binaryEBHeader))
	hashes = append(hashes, b.Header.BodyMR)
	merkle := BuildMerkleTreeStore(hashes)
	b.KeyMR = merkle[len(merkle)-1] // MerkleRoot is not marshalized in Dir Block

	return
}

func (b *DirectoryBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	newData = data

	fbh := new(DBlockHeader)
	newData, err = fbh.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}
	b.Header = fbh

	count := b.Header.BlockCount
	b.DBEntries = make([]*DBEntry, count)
	for i := uint32(0); i < count; i++ {
		b.DBEntries[i] = new(DBEntry)
		newData, err = b.DBEntries[i].UnmarshalBinaryData(newData)
		if err != nil {
			return
		}
	}

	return
}

func (b *DirectoryBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

func (b *DirectoryBlock) EncodableFields() map[string]reflect.Value {
	fields := map[string]reflect.Value{
		`Header`:    reflect.ValueOf(b.Header),
		`DBEntries`: reflect.ValueOf(b.DBEntries),
		`DBHash`:    reflect.ValueOf(b.DBHash),
	}
	return fields
}

func (b *DirBlockInfo) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, err = b.DBHash.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	data, err = b.BTCTxHash.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, b.BTCTxOffset)
	binary.Write(&buf, binary.BigEndian, b.BTCBlockHeight)

	/*	data, err = b.BTCBlockHash.MarshalBinary()
		if err != nil {
			return
		}
		buf.Write(data)
	*/
	data, err = b.DBMerkleRoot.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	// convert bool to one byte
	if b.BTCConfirmed {
		buf.Write([]byte{1})
	} else {
		buf.Write([]byte{0})
	}
	return buf.Bytes(), err
}

func (b *DirBlockInfo) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	newData = data

	b.DBHash = new(Hash)
	newData, err = b.DBHash.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.BTCTxHash = new(Hash)
	newData, err = b.BTCTxHash.UnmarshalBinaryData(newData)

	b.BTCTxOffset = int(binary.BigEndian.Uint32(newData[:4]))
	newData = newData[4:]

	b.BTCBlockHeight = int32(binary.BigEndian.Uint32(newData[:4]))
	newData = newData[4:]

	b.DBMerkleRoot = new(Hash)
	newData, err = b.DBMerkleRoot.UnmarshalBinaryData(newData)

	// convert one byte to bool
	if newData[0] > 0 {
		b.BTCConfirmed = true
	} else {
		b.BTCConfirmed = false
	}
	newData = newData[1:]

	return
}

func (b *DirBlockInfo) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}
