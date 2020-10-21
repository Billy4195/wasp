package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type ScEvent struct {
	MapObject
	code     int64
	contract string
	delay    int64
}

func (o *ScEvent) GetObjectId(keyId int32, typeId int32) int32 {
	return o.GetMapObjectId(keyId, typeId, map[int32]MapObjDesc{
		KeyParams: {OBJTYPE_MAP, func() WaspObject { return &ScEventParams{} }},
	})
}

func (o *ScEvent) Send() {
	o.vm.Trace("EVENT f'%s' c%d d%d a'%s'", o.vm.codeToFunc[int32(o.code)], o.code, o.delay, o.contract)
	if o.contract == "" {
		params := kv.NewMap()
		paramsId, ok := o.objects[KeyParams]
		if ok {
			params = o.vm.FindObject(paramsId).(*ScEventParams).Params
			params.ForEach(func(key kv.Key, value []byte) bool {
				o.vm.Trace("  PARAM '%s'", key)
				return true
			})
		}
		if params.IsEmpty() {
			params = nil
		}
		o.vm.ctx.SendRequestToSelfWithDelay(sctransaction.RequestCode(o.code), params, uint32(o.delay))
	}
	//TODO handle o.contract != ""
}

func (o *ScEvent) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.contract = ""
		o.code = 0
		o.delay = 0
	case KeyCode:
		o.code = value
	case KeyDelay:
		o.delay = value
	default:
		o.MapObject.SetInt(keyId, value)
	}
}

func (o *ScEvent) SetString(keyId int32, value string) {
	switch keyId {
	case KeyContract:
		o.contract = value
	case KeyFunction:
		code, ok := o.vm.funcToCode[value]
		if !ok {
			o.error("SetString: invalid function: %s", value)
			return
		}
		o.code = int64(code)
	default:
		o.MapObject.SetString(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScEvents struct {
	ArrayObject
}

func (a *ScEvents) GetObjectId(keyId int32, typeId int32) int32 {
	return a.GetArrayObjectId(keyId, typeId, func() WaspObject {
		event := &ScEvent{}
		event.name = "event"
		return event
	})
}

func (a *ScEvents) Send() {
	for i := 0; i < len(a.objects); i++ {
		request := a.vm.FindObject(a.objects[i]).(*ScEvent)
		request.Send()
	}
}

func (a *ScEvents) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		a.objects = nil
		return
	default:
		a.ArrayObject.SetInt(keyId, value)
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScEventParams struct {
	MapObject
	Params kv.Map
}

func (o *ScEventParams) InitVM(vm *wasmProcessor, keyId int32) {
	o.MapObject.InitVM(vm, keyId)
	o.Params = kv.NewMap()
}

func (o *ScEventParams) GetBytes(keyId int32) []byte {
	value, _ := o.Params.Get(kv.Key(o.vm.GetKey(keyId)))
	return value
}

func (o *ScEventParams) GetInt(keyId int32) int64 {
	value, ok, _ := o.Params.Codec().GetInt64(kv.Key(o.vm.GetKey(keyId)))
	if ok {
		return value
	}
	return o.MapObject.GetInt(keyId)
}

func (o *ScEventParams) GetObjectId(keyId int32, typeId int32) int32 {
	return o.MapObject.GetObjectId(keyId, typeId)
}

func (o *ScEventParams) GetString(keyId int32) string {
	value, ok, _ := o.Params.Codec().GetString(kv.Key(o.vm.GetKey(keyId)))
	if ok {
		return value
	}
	return o.MapObject.GetString(keyId)
}

func (o *ScEventParams) SetBytes(keyId int32, value []byte) {
	o.Params.Set(kv.Key(o.vm.GetKey(keyId)), value)
}

func (o *ScEventParams) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.Params = kv.NewMap()
	default:
		o.Params.Codec().SetInt64(kv.Key(o.vm.GetKey(keyId)), value)
	}
}

func (o *ScEventParams) SetString(keyId int32, value string) {
	o.Params.Codec().SetString(kv.Key(o.vm.GetKey(keyId)), value)
}