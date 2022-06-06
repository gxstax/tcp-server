package frame

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

type ReturnErrorWriter struct {
	W io.Writer
	Wn int		// 第几次调用 Write 返回错误
	wc int		// 写操作次数计数
}

func (w *ReturnErrorWriter) Write(p []byte) (n int, err error) {
	w.wc++
	if w.wc >= w.Wn {
		return 0, errors.New("write error")
	}
	return w.W.Write(p)
}

type ReturnErrorRead struct {
	R io.Reader
	Rn int		// 第几次调用 Read 返回错误
	rc int		// 读操作次数计数
}

func (r *ReturnErrorRead) Read(p []byte) (n int, err error) {
	r.rc++
	if r.rc >= r.Rn {
		return 0, errors.New("read error")
	}
	return r.R.Read(p)
}

func TestMyFrameCodec_Encode(t *testing.T) {
	codec := NewMyFrameCodec()
	buf := make([]byte, 0, 128)
	rw := bytes.NewBuffer(buf)

	err := codec.Encode(rw, []byte("hello"))
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
	}

	// 验证 Encode 的正确性
	var totalLen int32
	err = binary.Read(rw, binary.BigEndian, &totalLen)
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
	}

	if totalLen != 9 {
		t.Errorf("want 9 actual %d", totalLen)
	}

	left := rw.Bytes()
	if string(left) != "hello" {
		t.Errorf("want hello, actual %s", string(left))
	}
}

func TestMyFrameCodec_Decode(t *testing.T) {
	codec := NewMyFrameCodec()
	data := []byte{0x0, 0x0, 0x0, 0x9, 'h', 'e', 'l', 'l', 'o'}

	payload, err := codec.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
	}

	if string(payload) != "hello" {
		t.Errorf("want hello, actual %s", string(payload))
	}
}

func TestEncodeWithWriteFail(t *testing.T) {
	codec := NewMyFrameCodec()
	buf := make([]byte, 0, 128)
	w := bytes.NewBuffer(buf)

	// 模拟 nary.Write 返回错误
	err := codec.Encode(&ReturnErrorWriter{
		W: w,
		Wn: 1,
	}, []byte("hello"))
	if err == nil {
		t.Errorf("want non-nil, actural nil")
	}

	// 模拟 w.Write 返回错误
	err = codec.Encode(&ReturnErrorWriter{
		W: w,
		Wn: 2,
	}, []byte("hello"))
	if err == nil {
		t.Errorf("want non-nil, actual nil")
	}
}

func TestDecodeWithReadFail(t *testing.T) {
	codec := NewMyFrameCodec()
	data := []byte{0x0, 0x0, 0x0, 0x9, 'h', 'e', 'l', 'l', 'o'}

	// 模拟 binary.Read 返回错误
	_, err := codec.Decode(&ReturnErrorRead{
		R: bytes.NewReader(data),
		Rn: 1,
	})
	if err == nil {
		t.Errorf("want non-nil, actual nil")
	}

	// 模拟 io.Read 返回错误
	_, err = codec.Decode(&ReturnErrorRead{
		R: bytes.NewReader(data),
		Rn: 2,
	})
	if err == nil {
		t.Errorf("want non-nil, actual nil")
	}
}