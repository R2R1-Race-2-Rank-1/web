// Copyright (c) 2012-2018 Ugorji Nwoke. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

/*
MSGPACK

Msgpack-c implementation powers the c, c++, python, ruby, etc libraries.
We need to maintain compatibility with it and how it encodes integer values
without caring about the type.

For compatibility with behaviour of msgpack-c reference implementation:
  - Go intX (>0) and uintX
       IS ENCODED AS
    msgpack +ve fixnum, unsigned
  - Go intX (<0)
       IS ENCODED AS
    msgpack -ve fixnum, signed
*/

package codec

import (
	"fmt"
	"io"
	"math"
	"net/rpc"
	"reflect"
	"time"
)

const (
	mpPosFixNumMin byte = 0x00
	mpPosFixNumMax      = 0x7f
	mpFixMapMin         = 0x80
	mpFixMapMax         = 0x8f
	mpFixArrayMin       = 0x90
	mpFixArrayMax       = 0x9f
	mpFixStrMin         = 0xa0
	mpFixStrMax         = 0xbf
	mpNil               = 0xc0
	_                   = 0xc1
	mpFalse             = 0xc2
	mpTrue              = 0xc3
	mpFloat             = 0xca
	mpDouble            = 0xcb
	mpUint8             = 0xcc
	mpUint16            = 0xcd
	mpUint32            = 0xce
	mpUint64            = 0xcf
	mpInt8              = 0xd0
	mpInt16             = 0xd1
	mpInt32             = 0xd2
	mpInt64             = 0xd3

	// extensions below
	mpBin8     = 0xc4
	mpBin16    = 0xc5
	mpBin32    = 0xc6
	mpExt8     = 0xc7
	mpExt16    = 0xc8
	mpExt32    = 0xc9
	mpFixExt1  = 0xd4
	mpFixExt2  = 0xd5
	mpFixExt4  = 0xd6
	mpFixExt8  = 0xd7
	mpFixExt16 = 0xd8

	mpStr8  = 0xd9 // new
	mpStr16 = 0xda
	mpStr32 = 0xdb

	mpArray16 = 0xdc
	mpArray32 = 0xdd

	mpMap16 = 0xde
	mpMap32 = 0xdf

	mpNegFixNumMin = 0xe0
	mpNegFixNumMax = 0xff
)

var mpTimeExtTag int8 = -1
var mpTimeExtTagU = uint8(mpTimeExtTag)

// MsgpackSpecRpcMultiArgs is a special type which signifies to the MsgpackSpecRpcCodec
// that the backend RPC service takes multiple arguments, which have been arranged
// in sequence in the slice.
//
// The Codec then passes it AS-IS to the rpc service (without wrapping it in an
// array of 1 element).
type MsgpackSpecRpcMultiArgs []interface{}

// A MsgpackContainer type specifies the different types of msgpackContainers.
type msgpackContainerType struct {
	fixCutoff                   int
	bFixMin, b8, b16, b32       byte
	hasFixMin, has8, has8Always bool
}

var (
	msgpackContainerStr = msgpackContainerType{
		32, mpFixStrMin, mpStr8, mpStr16, mpStr32, true, true, false,
	}
	msgpackContainerBin = msgpackContainerType{
		0, 0, mpBin8, mpBin16, mpBin32, false, true, true,
	}
	msgpackContainerList = msgpackContainerType{
		16, mpFixArrayMin, 0, mpArray16, mpArray32, true, false, false,
	}
	msgpackContainerMap = msgpackContainerType{
		16, mpFixMapMin, 0, mpMap16, mpMap32, true, false, false,
	}
)

//---------------------------------------------

type msgpackEncDriver struct {
	noBuiltInTypes
	encDriverNoopContainerWriter
	// encNoSeparator
	e *Encoder
	w encWriter
	h *MsgpackHandle
	x [8]byte
	_ [3]uint64 // padding
}

func (e *msgpackEncDriver) EncodeNil() {
	e.w.writen1(mpNil)
}

func (e *msgpackEncDriver) EncodeInt(i int64) {
	// if i >= 0 {
	// 	e.EncodeUint(uint64(i))
	// } else if false &&
	if i > math.MaxInt8 {
		if i <= math.MaxInt16 {
			e.w.writen1(mpInt16)
			bigenHelper{e.x[:2], e.w}.writeUint16(uint16(i))
		} else if i <= math.MaxInt32 {
			e.w.writen1(mpInt32)
			bigenHelper{e.x[:4], e.w}.writeUint32(uint32(i))
		} else {
			e.w.writen1(mpInt64)
			bigenHelper{e.x[:8], e.w}.writeUint64(uint64(i))
		}
	} else if i >= -32 {
		if e.h.NoFixedNum {
			e.w.writen2(mpInt8, byte(i))
		} else {
			e.w.writen1(byte(i))
		}
	} else if i >= math.MinInt8 {
		e.w.writen2(mpInt8, byte(i))
	} else if i >= math.MinInt16 {
		e.w.writen1(mpInt16)
		bigenHelper{e.x[:2], e.w}.writeUint16(uint16(i))
	} else if i >= math.MinInt32 {
		e.w.writen1(mpInt32)
		bigenHelper{e.x[:4], e.w}.writeUint32(uint32(i))
	} else {
		e.w.writen1(mpInt64)
		bigenHelper{e.x[:8], e.w}.writeUint64(uint64(i))
	}
}

func (e *msgpackEncDriver) EncodeUint(i uint64) {
	if i <= math.MaxInt8 {
		if e.h.NoFixedNum {
			e.w.writen2(mpUint8, byte(i))
		} else {
			e.w.writen1(byte(i))
		}
	} else if i <= math.MaxUint8 {
		e.w.writen2(mpUint8, byte(i))
	} else if i <= math.MaxUint16 {
		e.w.writen1(mpUint16)
		bigenHelper{e.x[:2], e.w}.writeUint16(uint16(i))
	} else if i <= math.MaxUint32 {
		e.w.writen1(mpUint32)
		bigenHelper{e.x[:4], e.w}.writeUint32(uint32(i))
	} else {
		e.w.writen1(mpUint64)
		bigenHelper{e.x[:8], e.w}.writeUint64(uint64(i))
	}
}

func (e *msgpackEncDriver) EncodeBool(b bool) {
	if b {
		e.w.writen1(mpTrue)
	} else {
		e.w.writen1(mpFalse)
	}
}

func (e *msgpackEncDriver) EncodeFloat32(f float32) {
	e.w.writen1(mpFloat)
	bigenHelper{e.x[:4], e.w}.writeUint32(math.Float32bits(f))
}

func (e *msgpackEncDriver) EncodeFloat64(f float64) {
	e.w.writen1(mpDouble)
	bigenHelper{e.x[:8], e.w}.writeUint64(math.Float64bits(f))
}

func (e *msgpackEncDriver) EncodeTime(t time.Time) {
	if t.IsZero() {
		e.EncodeNil()
		return
	}
	t = t.UTC()
	sec, nsec := t.Unix(), uint64(t.Nanosecond())
	var data64 uint64
	var l = 4
	if sec >= 0 && sec>>34 == 0 {
		data64 = (nsec << 34) | uint64(sec)
		if data64&0xffffffff00000000 != 0 {
			l = 8
		}
	} else {
		l = 12
	}
	if e.h.WriteExt {
		e.encodeExtPreamble(mpTimeExtTagU, l)
	} else {
		e.writeContainerLen(msgpackContainerStr, l)
	}
	switch l {
	case 4:
		bigenHelper{e.x[:4], e.w}.writeUint32(uint32(data64))
	case 8:
		bigenHelper{e.x[:8], e.w}.writeUint64(data64)
	case 12:
		bigenHelper{e.x[:4], e.w}.writeUint32(uint32(nsec))
		bigenHelper{e.x[:8], e.w}.writeUint64(uint64(sec))
	}
}

func (e *msgpackEncDriver) EncodeExt(v interface{}, xtag uint64, ext Ext, _ *Encoder) {
	bs := ext.WriteExt(v)
	if bs == nil {
		e.EncodeNil()
		return
	}
	if e.h.WriteExt {
		e.encodeExtPreamble(uint8(xtag), len(bs))
		e.w.writeb(bs)
	} else {
		e.EncodeStringBytes(cRAW, bs)
	}
}

func (e *msgpackEncDriver) EncodeRawExt(re *RawExt, _ *Encoder) {
	e.encodeExtPreamble(uint8(re.Tag), len(re.Data))
	e.w.writeb(re.Data)
}

func (e *msgpackEncDriver) encodeExtPreamble(xtag byte, l int) {
	if l == 1 {
		e.w.writen2(mpFixExt1, xtag)
	} else if l == 2 {
		e.w.writen2(mpFixExt2, xtag)
	} else if l == 4 {
		e.w.writen2(mpFixExt4, xtag)
	} else if l == 8 {
		e.w.writen2(mpFixExt8, xtag)
	} else if l == 16 {
		e.w.writen2(mpFixExt16, xtag)
	} else if l < 256 {
		e.w.writen2(mpExt8, byte(l))
		e.w.writen1(xtag)
	} else if l < 65536 {
		e.w.writen1(mpExt16)
		bigenHelper{e.x[:2], e.w}.writeUint16(uint16(l))
		e.w.writen1(xtag)
	} else {
		e.w.writen1(mpExt32)
		bigenHelper{e.x[:4], e.w}.writeUint32(uint32(l))
		e.w.writen1(xtag)
	}
}

func (e *msgpackEncDriver) WriteArrayStart(length int) {
	e.writeContainerLen(msgpackContainerList, length)
}

func (e *msgpackEncDriver) WriteMapStart(length int) {
	e.writeContainerLen(msgpackContainerMap, length)
}

func (e *msgpackEncDriver) EncodeString(c charEncoding, s string) {
	slen := len(s)
	if c == cRAW && e.h.WriteExt {
		e.writeContainerLen(msgpackContainerBin, slen)
	} else {
		e.writeContainerLen(msgpackContainerStr, slen)
	}
	if slen > 0 {
		e.w.writestr(s)
	}
}

func (e *msgpackEncDriver) EncodeStringBytes(c charEncoding, bs []byte) {
	if bs == nil {
		e.EncodeNil()
		return
	}
	slen := len(bs)
	if c == cRAW && e.h.WriteExt {
		e.writeContainerLen(msgpackContainerBin, slen)
	} else {
		e.writeContainerLen(msgpackContainerStr, slen)
	}
	if slen > 0 {
		e.w.writeb(bs)
	}
}

func (e *msgpackEncDriver) writeContainerLen(ct msgpackContainerType, l int) {
	if ct.hasFixMin && l < ct.fixCutoff {
		e.w.writen1(ct.bFixMin | byte(l))
	} else if ct.has8 && l < 256 && (ct.has8Always || e.h.WriteExt) {
		e.w.writen2(ct.b8, uint8(l))
	} else if l < 65536 {
		e.w.writen1(ct.b16)
		bigenHelper{e.x[:2], e.w}.writeUint16(uint16(l))
	} else {
		e.w.writen1(ct.b32)
		bigenHelper{e.x[:4], e.w}.writeUint32(uint32(l))
	}
}

//---------------------------------------------

type msgpackDecDriver struct {
	d *Decoder
	r decReader // *Decoder decReader decReaderT
	h *MsgpackHandle
	// b      [scratchByteArrayLen]byte
	bd     byte
	bdRead bool
	br     bool // bytes reader
	noBuiltInTypes
	// noStreamingCodec
	// decNoSeparator
	decDriverNoopContainerReader
	_ [3]uint64 // padding
}

// Note: This returns either a primitive (int, bool, etc) for non-containers,
// or a containerType, or a specific type denoting nil or extension.
// It is called when a nil interface{} is passed, leaving it up to the DecDriver
// to introspect the stream and decide how best to decode.
// It deciphers the value by looking at the stream first.
func (d *msgpackDecDriver) DecodeNaked() {
	if !d.bdRead {
		d.readNextBd()
	}
	bd := d.bd
	n := d.d.n
	var decodeFurther bool

	switch bd {
	case mpNil:
		n.v = valueTypeNil
		d.bdRead = false
	case mpFalse:
		n.v = valueTypeBool
		n.b = false
	case mpTrue:
		n.v = valueTypeBool
		n.b = true

	case mpFloat:
		n.v = valueTypeFloat
		n.f = float64(math.Float32frombits(bigen.Uint32(d.r.readx(4))))
	case mpDouble:
		n.v = valueTypeFloat
		n.f = math.Float64frombits(bigen.Uint64(d.r.readx(8)))

	case mpUint8:
		n.v = valueTypeUint
		n.u = uint64(d.r.readn1())
	case mpUint16:
		n.v = valueTypeUint
		n.u = uint64(bigen.Uint16(d.r.readx(2)))
	case mpUint32:
		n.v = valueTypeUint
		n.u = uint64(bigen.Uint32(d.r.readx(4)))
	case mpUint64:
		n.v = valueTypeUint
		n.u = uint64(bigen.Uint64(d.r.readx(8)))

	case mpInt8:
		n.v = valueTypeInt
		n.i = int64(int8(d.r.readn1()))
	case mpInt16:
		n.v = valueTypeInt
		n.i = int64(int16(bigen.Uint16(d.r.readx(2))))
	case mpInt32:
		n.v = valueTypeInt
		n.i = int64(int32(bigen.Uint32(d.r.readx(4))))
	case mpInt64:
		n.v = valueTypeInt
		n.i = int64(int64(bigen.Uint64(d.r.readx(8))))

	default:
		switch {
		case bd >= mpPosFixNumMin && bd <= mpPosFixNumMax:
			// positive fixnum (always signed)
			n.v = valueTypeInt
			n.i = int64(int8(bd))
		case bd >= mpNegFixNumMin && bd <= mpNegFixNumMax:
			// negative fixnum
			n.v = valueTypeInt
			n.i = int64(int8(bd))
		case bd == mpStr8, bd == mpStr16, bd == mpStr32, bd >= mpFixStrMin && bd <= mpFixStrMax:
			if d.h.RawToString {
				n.v = valueTypeString
				n.s = d.DecodeString()
			} else {
				n.v = valueTypeBytes
				n.l = d.DecodeBytes(nil, false)
			}
		case bd == mpBin8, bd == mpBin16, bd == mpBin32:
			n.v = valueTypeBytes
			n.l = d.DecodeBytes(nil, false)
		case bd == mpArray16, bd == mpArray32, bd >= mpFixArrayMin && bd <= mpFixArrayMax:
			n.v = valueTypeArray
			decodeFurther = true
		case bd == mpMap16, bd == mpMap32, bd >= mpFixMapMin && bd <= mpFixMapMax:
			n.v = valueTypeMap
			decodeFurther = true
		case bd >= mpFixExt1 && bd <= mpFixExt16, bd >= mpExt8 && bd <= mpExt32:
			n.v = valueTypeExt
			clen := d.readExtLen()
			n.u = uint64(d.r.readn1())
			if n.u == uint64(mpTimeExtTagU) {
				n.v = valueTypeTime
				n.t = d.decodeTime(clen)
			} else {
				n.l = d.r.readx(clen)
			}
		default:
			d.d.errorf("Nil-Deciphered DecodeValue: %s: hex: %x, dec: %d", msgBadDesc, bd, bd)
		}
	}
	if !decodeFurther {
		d.bdRead = false
	}
	if n.v == valueTypeUint && d.h.SignedInteger {
		n.v = valueTypeInt
		n.i = int64(n.u)
	}
	return
}

// int can be decoded from msgpack type: intXXX or uintXXX
func (d *msgpackDecDriver) DecodeInt64() (i int64) {
	if !d.bdRead {
		d.readNextBd()
	}
	switch d.bd {
	case mpUint8:
		i = int64(uint64(d.r.readn1()))
	case mpUint16:
		i = int64(uint64(bigen.Uint16(d.r.readx(2))))
	case mpUint32:
		i = int64(uint64(bigen.Uint32(d.r.readx(4))))
	case mpUint64:
		i = int64(bigen.Uint64(d.r.readx(8)))
	case mpInt8:
		i = int64(int8(d.r.readn1()))
	case mpInt16:
		i = int64(int16(bigen.Uint16(d.r.readx(2))))
	case mpInt32:
		i = int64(int32(bigen.Uint32(d.r.readx(4))))
	case mpInt64:
		i = int64(bigen.Uint64(d.r.readx(8)))
	default:
		switch {
		case d.bd >= mpPosFixNumMin && d.bd <= mpPosFixNumMax:
			i = int64(int8(d.bd))
		case d.bd >= mpNegFixNumMin && d.bd <= mpNegFixNumMax:
			i = int64(int8(d.bd))
		default:
			d.d.errorf("Unhandled single-byte unsigned integer value: %s: %x", msgBadDesc, d.bd)
			return
		}
	}
	d.bdRead = false
	return
}

// uint can be decoded from msgpack type: intXXX or uintXXX
func (d *msgpackDecDriver) DecodeUint64() (ui uint64) {
	if !d.bdRead {
		d.readNextBd()
	}
	switch d.bd {
	case mpUint8:
		ui = uint64(d.r.readn1())
	case mpUint16:
		ui = uint64(bigen.Uint16(d.r.readx(2)))
	case mpUint32:
		ui = uint64(bigen.Uint32(d.r.readx(4)))
	case mpUint64:
		ui = bigen.Uint64(d.r.readx(8))
	case mpInt8:
		if i := int64(int8(d.r.readn1())); i >= 0 {
			ui = uint64(i)
		} else {
			d.d.errorf("Assigning negative signed value: %v, to unsigned type", i)
			return
		}
	case mpInt16:
		if i := int64(int16(bigen.Uint16(d.r.readx(2)))); i >= 0 {
			ui = uint64(i)
		} else {
			d.d.errorf("Assigning negative signed value: %v, to unsigned type", i)
			return
		}
	case mpInt32:
		if i := int64(int32(bigen.Uint32(d.r.readx(4)))); i >= 0 {
			ui = uint64(i)
		} else {
			d.d.errorf("Assigning negative signed value: %v, to unsigned type", i)
			return
		}
	case mpInt64:
		if i := int64(bigen.Uint64(d.r.readx(8))); i >= 0 {
			ui = uint64(i)
		} else {
			d.d.errorf("Assigning negative signed value: %v, to unsigned type", i)
			return
		}
	default:
		switch {
		case d.bd >= mpPosFixNumMin && d.bd <= mpPosFixNumMax:
			ui = uint64(d.bd)
		case d.bd >= mpNegFixNumMin && d.bd <= mpNegFixNumMax:
			d.d.errorf("Assigning negative signed value: %v, to unsigned type", int(d.bd))
			return
		default:
			d.d.errorf("Unhandled single-byte unsigned integer value: %s: %x", msgBadDesc, d.bd)
			return
		}
	}
	d.bdRead = false
	return
}

// float can either be decoded from msgpack type: float, double or intX
func (d *msgpackDecDriver) DecodeFloat64() (f float64) {
	if !d.bdRead {
		d.readNextBd()
	}
	if d.bd == mpFloat {
		f = float64(math.Float32frombits(bigen.Uint32(d.r.readx(4))))
	} else if d.bd == mpDouble {
		f = math.Float64frombits(bigen.Uint64(d.r.readx(8)))
	} else {
		f = float64(d.DecodeInt64())
	}
	d.bdRead = false
	return
}

// bool can be decoded from bool, fixnum 0 or 1.
func (d *msgpackDecDriver) DecodeBool() (b bool) {
	if !d.bdRead {
		d.readNextBd()
	}
	if d.bd == mpFalse || d.bd == 0 {
		// b = false
	} else if d.bd == mpTrue || d.bd == 1 {
		b = true
	} else {
		d.d.errorf("Invalid single-byte value for bool: %s: %x", msgBadDesc, d.bd)
		return
	}
	d.bdRead = false
	return
}

func (d *msgpackDecDriver) DecodeBytes(bs []byte, zerocopy bool) (bsOut []byte) {
	if !d.bdRead {
		d.readNextBd()
	}

	// check if an "array" of uint8's (see ContainerType for how to infer if an array)
	bd := d.bd
	// DecodeBytes could be from: bin str fixstr fixarray array ...
	var clen int
	vt := d.ContainerType()
	switch vt {
	case valueTypeBytes:
		// valueTypeBytes may be a mpBin or an mpStr container
		if bd == mpBin8 || bd == mpBin16 || bd == mpBin32 {
			clen = d.readContainerLen(msgpackContainerBin)
		} else {
			clen = d.readContainerLen(msgpackContainerStr)
		}
	case valueTypeString:
		clen = d.readContainerLen(msgpackContainerStr)
	case valueTypeArray:
		if zerocopy && len(bs) == 0 {
			bs = d.d.b[:]
		}
		bsOut, _ = fastpathTV.DecSliceUint8V(bs, true, d.d)
		return
	default:
		d.d.errorf("invalid container type: expecting bin|str|array, got: 0x%x", uint8(vt))
		return
	}

	// these are (bin|str)(8|16|32)
	d.bdRead = false
	// bytes may be nil, so handle it. if nil, clen=-1.
	if clen < 0 {
		return nil
	}
	if zerocopy {
		if d.br {
			return d.r.readx(clen)
		} else if len(bs) == 0 {
			bs = d.d.b[:]
		}
	}
	return decByteSlice(d.r, clen, d.h.MaxInitLen, bs)
}

func (d *msgpackDecDriver) DecodeString() (s string) {
	return string(d.DecodeBytes(d.d.b[:], true))
}

func (d *msgpackDecDriver) DecodeStringAsBytes() (s []byte) {
	return d.DecodeBytes(d.d.b[:], true)
}

func (d *msgpackDecDriver) readNextBd() {
	d.bd = d.r.readn1()
	d.bdRead = true
}

func (d *msgpackDecDriver) uncacheRead() {
	if d.bdRead {
		d.r.unreadn1()
		d.bdRead = false
	}
}

func (d *msgpackDecDriver) ContainerType() (vt valueType) {
	if !d.bdRead {
		d.readNextBd()
	}
	bd := d.bd
	if bd == mpNil {
		return valueTypeNil
	} else if bd == mpBin8 || bd == mpBin16 || bd == mpBin32 ||
		(!d.h.RawToString &&
			(bd == mpStr8 || bd == mpStr16 || bd == mpStr32 || (bd >= mpFixStrMin && bd <= mpFixStrMax))) {
		return valueTypeBytes
	} else if d.h.RawToString &&
		(bd == mpStr8 || bd == mpStr16 || bd == mpStr32 || (bd >= mpFixStrMin && bd <= mpFixStrMax)) {
		return valueTypeString
	} else if bd == mpArray16 || bd == mpArray32 || (bd >= mpFixArrayMin && bd <= mpFixArrayMax) {
		return valueTypeArray
	} else if bd == mpMap16 || bd == mpMap32 || (bd >= mpFixMapMin && bd <= mpFixMapMax) {
		return valueTypeMap
	}
	// else {
	// d.d.errorf("isContainerType: unsupported parameter: %v", vt)
	// }
	return valueTypeUnset
}

func (d *msgpackDecDriver) TryDecodeAsNil() (v bool) {
	if !d.bdRead {
		d.readNextBd()
	}
	if d.bd == mpNil {
		d.bdRead = false
		return true
	}
	return
}

func (d *msgpackDecDriver) readContainerLen(ct msgpackContainerType) (clen int) {
	bd := d.bd
	if bd == mpNil {
		clen = -1 // to represent nil
	} else if bd == ct.b8 {
		clen = int(d.r.readn1())
	} else if bd == ct.b16 {
		clen = int(bigen.Uint16(d.r.readx(2)))
	} else if bd == ct.b32 {
		clen = int(bigen.Uint32(d.r.readx(4)))
	} else if (ct.bFixMin & bd) == ct.bFixMin {
		clen = int(ct.bFixMin ^ bd)
	} else {
		d.d.errorf("readContainerLen: %s: hex: %x, decimal: %d", msgBadDesc, bd, bd)
		return
	}
	d.bdRead = false
	return
}

func (d *msgpackDecDriver) ReadMapStart() int {
	if !d.bdRead {
		d.readNextBd()
	}
	return d.readContainerLen(msgpackContainerMap)
}

func (d *msgpackDecDriver) ReadArrayStart() int {
	if !d.bdRead {
		d.readNextBd()
	}
	return d.readContainerLen(msgpackContainerList)
}

func (d *msgpackDecDriver) readExtLen() (clen int) {
	switch d.bd {
	case mpNil:
		clen = -1 // to represent nil
	case mpFixExt1:
		clen = 1
	case mpFixExt2:
		clen = 2
	case mpFixExt4:
		clen = 4
	case mpFixExt8:
		clen = 8
	case mpFixExt16:
		clen = 16
	case mpExt8:
		clen = int(d.r.readn1())
	case mpExt16:
		clen = int(bigen.Uint16(d.r.readx(2)))
	case mpExt32:
		clen = int(bigen.Uint32(d.r.readx(4)))
	default:
		d.d.errorf("decoding ext bytes: found unexpected byte: %x", d.bd)
		return
	}
	return
}

func (d *msgpackDecDriver) DecodeTime() (t time.Time) {
	// decode time from string bytes or ext
	if !d.bdRead {
		d.readNextBd()
	}
	if d.bd == mpNil {
		d.bdRead = false
		return
	}
	var clen int
	switch d.ContainerType() {
	case valueTypeBytes, valueTypeString:
		clen = d.readContainerLen(msgpackContainerStr)
	default:
		// expect to see mpFixExt4,-1 OR mpFixExt8,-1 OR mpExt8,12,-1
		d.bdRead = false
		b2 := d.r.readn1()
		if d.bd == mpFixExt4 && b2 == mpTimeExtTagU {
			clen = 4
		} else if d.bd == mpFixExt8 && b2 == mpTimeExtTagU {
			clen = 8
		} else if d.bd == mpExt8 && b2 == 12 && d.r.readn1() == mpTimeExtTagU {
			clen = 12
		} else {
			d.d.errorf("invalid bytes for decoding time as extension: got 0x%x, 0x%x", d.bd, b2)
			return
		}
	}
	return d.decodeTime(clen)
}

func (d *msgpackDecDriver) decodeTime(clen int) (t time.Time) {
	// bs = d.r.readx(clen)
	d.bdRead = false
	switch clen {
	case 4:
		t = time.Unix(int64(bigen.Uint32(d.r.readx(4))), 0).UTC()
	case 8:
		tv := bigen.Uint64(d.r.readx(8))
		t = time.Unix(int64(tv&0x00000003ffffffff), int64(tv>>34)).UTC()
	case 12:
		nsec := bigen.Uint32(d.r.readx(4))
		sec := bigen.Uint64(d.r.readx(8))
		t = time.Unix(int64(sec), int64(nsec)).UTC()
	default:
		d.d.errorf("invalid length of bytes for decoding time - expecting 4 or 8 or 12, got %d", clen)
		return
	}
	return
}

func (d *msgpackDecDriver) DecodeExt(rv interface{}, xtag uint64, ext Ext) (realxtag uint64) {
	if xtag > 0xff {
		d.d.errorf("decodeExt: tag must be <= 0xff; got: %v", xtag)
		return
	}
	realxtag1, xbs := d.decodeExtV(ext != nil, uint8(xtag))
	realxtag = uint64(realxtag1)
	if ext == nil {
		re := rv.(*RawExt)
		re.Tag = realxtag
		re.Data = detachZeroCopyBytes(d.br, re.Data, xbs)
	} else {
		ext.ReadExt(rv, xbs)
	}
	return
}

func (d *msgpackDecDriver) decodeExtV(verifyTag bool, tag byte) (xtag byte, xbs []byte) {
	if !d.bdRead {
		d.readNextBd()
	}
	xbd := d.bd
	if xbd == mpBin8 || xbd == mpBin16 || xbd == mpBin32 {
		xbs = d.DecodeBytes(nil, true)
	} else if xbd == mpStr8 || xbd == mpStr16 || xbd == mpStr32 ||
		(xbd >= mpFixStrMin && xbd <= mpFixStrMax) {
		xbs = d.DecodeStringAsBytes()
	} else {
		clen := d.readExtLen()
		xtag = d.r.readn1()
		if verifyTag && xtag != tag {
			d.d.errorf("Wrong extension tag. Got %b. Expecting: %v", xtag, tag)
			return
		}
		xbs = d.r.readx(clen)
	}
	d.bdRead = false
	return
}

//--------------------------------------------------

//MsgpackHandle is a Handle for the Msgpack Schema-Free Encoding Format.
type MsgpackHandle struct {
	BasicHandle

	// RawToString controls how raw bytes are decoded into a nil interface{}.
	RawToString bool

	// NoFixedNum says to output all signed integers as 2-bytes, never as 1-byte fixednum.
	NoFixedNum bool

	// WriteExt flag supports encoding configured extensions with extension tags.
	// It also controls whether other elements of the new spec are encoded (ie Str8).
	//
	// With WriteExt=false, configured extensions are serialized as raw bytes
	// and Str8 is not encoded.
	//
	// A stream can still be decoded into a typed value, provided an appropriate value
	// is provided, but the type cannot be inferred from the stream. If no appropriate
	// type is provided (e.g. decoding into a nil interface{}), you get back
	// a []byte or string based on the setting of RawToString.
	WriteExt bool

	binaryEncodingType
	noElemSeparators

	_ [1]uint64 // padding
}

// Name returns the name of the handle: msgpack
func (h *MsgpackHandle) Name() string { return "msgpack" }

// SetBytesExt sets an extension
func (h *MsgpackHandle) SetBytesExt(rt reflect.Type, tag uint64, ext BytesExt) (err error) {
	return h.SetExt(rt, tag, &extWrapper{ext, interfaceExtFailer{}})
}

func (h *MsgpackHandle) newEncDriver(e *Encoder) encDriver {
	return &msgpackEncDriver{e: e, w: e.w, h: h}
}

func (h *MsgpackHandle) newDecDriver(d *Decoder) decDriver {
	return &msgpackDecDriver{d: d, h: h, r: d.r, br: d.bytes}
}

func (e *msgpackEncDriver) reset() {
	e.w = e.e.w
}

func (d *msgpackDecDriver) reset() {
	d.r, d.br = d.d.r, d.d.bytes
	d.bd, d.bdRead = 0, false
}

//--------------------------------------------------

type msgpackSpecRpcCodec struct {
	rpcCodec
}

// /////////////// Spec RPC Codec ///////////////////
func (c *msgpackSpecRpcCodec) WriteRequest(r *rpc.Request, body interface{}) error {
	// WriteRequest can write to both a Go service, and other services that do
	// not abide by the 1 argument rule of a Go service.
	// We discriminate based on if the body is a MsgpackSpecRpcMultiArgs
	var bodyArr []interface{}
	if m, ok := body.(MsgpackSpecRpcMultiArgs); ok {
		bodyArr = ([]interface{})(m)
	} else {
		bodyArr = []interface{}{body}
	}
	r2 := []interface{}{0, uint32(r.Seq), r.ServiceMethod, bodyArr}
	return c.write(r2, nil, false)
}

func (c *msgpackSpecRpcCodec) WriteResponse(r *rpc.Response, body interface{}) error {
	var moe interface{}
	if r.Error != "" {
		moe = r.Error
	}
	if moe != nil && body != nil {
		body = nil
	}
	r2 := []interface{}{1, uint32(r.Seq), moe, body}
	return c.write(r2, nil, false)
}

func (c *msgpackSpecRpcCodec) ReadResponseHeader(r *rpc.Response) error {
	return c.parseCustomHeader(1, &r.Seq, &r.Error)
}

func (c *msgpackSpecRpcCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.parseCustomHeader(0, &r.Seq, &r.ServiceMethod)
}

func (c *msgpackSpecRpcCodec) ReadRequestBody(body interface{}) error {
	if body == nil { // read and discard
		return c.read(nil)
	}
	bodyArr := []interface{}{body}
	return c.read(&bodyArr)
}

func (c *msgpackSpecRpcCodec) parseCustomHeader(expectTypeByte byte, msgid *uint64, methodOrError *string) (err error) {
	if c.isClosed() {
		return io.EOF
	}

	// We read the response header by hand
	// so that the body can be decoded on its own from the stream at a later time.

	const fia byte = 0x94 //four item array descriptor value
	// Not sure why the panic of EOF is swallowed above.
	// if bs1 := c.dec.r.readn1(); bs1 != fia {
	// 	err = fmt.Errorf("Unexpected value for array descriptor: Expecting %v. Received %v", fia, bs1)
	// 	return
	// }
	var ba [1]byte
	var n int
	for {
		n, err = c.r.Read(ba[:])
		if err != nil {
			return
		}
		if n == 1 {
			break
		}
	}

	var b = ba[0]
	if b != fia {
		err = fmt.Errorf("Unexpected value for array descriptor: Expecting %v. Received %v", fia, b)
	} else {
		err = c.read(&b)
		if err == nil {
			if b != expectTypeByte {
				err = fmt.Errorf("Unexpected byte descriptor. Expecting %v; Received %v",
					expectTypeByte, b)
			} else {
				err = c.read(msgid)
				if err == nil {
					err = c.read(methodOrError)
				}
			}
		}
	}
	return
}

//--------------------------------------------------

// msgpackSpecRpc is the implementation of Rpc that uses custom communication protocol
// as defined in the msgpack spec at https://github.com/msgpack-rpc/msgpack-rpc/blob/master/spec.md
type msgpackSpecRpc struct{}

// MsgpackSpecRpc implements Rpc using the communication protocol defined in
// the msgpack spec at https://github.com/msgpack-rpc/msgpack-rpc/blob/master/spec.md .
//
// See GoRpc documentation, for information on buffering for better performance.
var MsgpackSpecRpc msgpackSpecRpc

func (x msgpackSpecRpc) ServerCodec(conn io.ReadWriteCloser, h Handle) rpc.ServerCodec {
	return &msgpackSpecRpcCodec{newRPCCodec(conn, h)}
}

func (x msgpackSpecRpc) ClientCodec(conn io.ReadWriteCloser, h Handle) rpc.ClientCodec {
	return &msgpackSpecRpcCodec{newRPCCodec(conn, h)}
}

var _ decDriver = (*msgpackDecDriver)(nil)
var _ encDriver = (*msgpackEncDriver)(nil)
