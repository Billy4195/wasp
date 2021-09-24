// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

// @formatter:off

#![allow(dead_code)]

use wasmlib::*;
use wasmlib::host::*;

pub type ImmutableAllowancesForAgent = MapAgentIDToImmutableInt64;

pub struct MapAgentIDToImmutableInt64 {
    pub(crate) obj_id: i32,
}

impl MapAgentIDToImmutableInt64 {
    pub fn get_int64(&self, key: &ScAgentID) -> ScImmutableInt64 {
        ScImmutableInt64::new(self.obj_id, key.get_key_id())
    }
}

pub type MutableAllowancesForAgent = MapAgentIDToMutableInt64;

pub struct MapAgentIDToMutableInt64 {
    pub(crate) obj_id: i32,
}

impl MapAgentIDToMutableInt64 {
    pub fn clear(&self) {
        clear(self.obj_id)
    }

    pub fn get_int64(&self, key: &ScAgentID) -> ScMutableInt64 {
        ScMutableInt64::new(self.obj_id, key.get_key_id())
    }
}

// @formatter:on
