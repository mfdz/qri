// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package cron_fbs

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type ShellScriptOptions struct {
	_tab flatbuffers.Table
}

func GetRootAsShellScriptOptions(buf []byte, offset flatbuffers.UOffsetT) *ShellScriptOptions {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ShellScriptOptions{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *ShellScriptOptions) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *ShellScriptOptions) Table() flatbuffers.Table {
	return rcv._tab
}

func ShellScriptOptionsStart(builder *flatbuffers.Builder) {
	builder.StartObject(0)
}
func ShellScriptOptionsEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}