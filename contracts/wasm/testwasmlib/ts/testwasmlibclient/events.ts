// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

import * as wasmclient from "wasmclient"

const testWasmLibHandlers = new Map<string, (evt: TestWasmLibEvents, msg: string[]) => void>([
	["testwasmlib.test", (evt: TestWasmLibEvents, msg: string[]) => evt.test(new EventTest(msg))],
]);

export class TestWasmLibEvents implements wasmclient.IEventHandler {
	test: (EventTest) => void = () => {};

	public callHandler(topic: string, params: string[]): void {
		const handler = testWasmLibHandlers.get(topic);
		if (handler !== undefined) {
			handler(this, params);
		}
	}

	public onTestWasmLibTest(handler: (EventTest) => void): void {
		this.test = handler;
	}
}

export class EventTest extends wasmclient.Event {
	public readonly address: wasmclient.Address;
	public readonly name: wasmclient.String;
	
	public constructor(msg: string[]) {
		super(msg)
		this.address = this.nextAddress();
		this.name = this.nextString();
	}
}