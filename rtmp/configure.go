// Copyright © 2021 Kris Nóva <kris@nivenly.com>
// Copyright (c) 2017 吴浩麟
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// ────────────────────────────────────────────────────────────────────────────
//
//  ████████╗██╗    ██╗██╗███╗   ██╗██╗  ██╗
//  ╚══██╔══╝██║    ██║██║████╗  ██║╚██╗██╔╝
//     ██║   ██║ █╗ ██║██║██╔██╗ ██║ ╚███╔╝
//     ██║   ██║███╗██║██║██║╚██╗██║ ██╔██╗
//     ██║   ╚███╔███╔╝██║██║ ╚████║██╔╝ ██╗
//     ╚═╝    ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝
//
// ────────────────────────────────────────────────────────────────────────────

package rtmp

import (
	"fmt"

	"github.com/kris-nova/logger"

	"github.com/gwuhaolin/livego/utils/uid"
	"github.com/patrickmn/go-cache"
)

/*
{
  "server": [
    {
      "appname": "live",
      "live": true,
	  "hls": true,
	  "static_push": []
    }
  ]
}
*/

type Application struct {
	Appname    string   `mapstructure:"appname"`
	Live       bool     `mapstructure:"live"`
	Hls        bool     `mapstructure:"hls"`
	Flv        bool     `mapstructure:"flv"`
	Api        bool     `mapstructure:"api"`
	StaticPush []string `mapstructure:"static_push"`
}

type JWT struct {
	Secret    string `mapstructure:"secret"`
	Algorithm string `mapstructure:"algorithm"`
}

func CheckAppName(appname string) bool {

	// We no longer block Apps based on name.
	return true
}

func GetStaticPushUrlList(appname string) ([]string, bool) {

	// We no longer block static push
	return nil, true
}

type RoomKeysType struct {
	localCache *cache.Cache
}

var RoomKeys = &RoomKeysType{
	localCache: cache.New(cache.NoExpiration, 0),
}

func (r *RoomKeysType) SetKey(channel string) (key string, err error) {

	for {
		key = uid.RandStringRunes(48)
		if _, found := r.localCache.Get(key); !found {
			r.localCache.SetDefault(channel, key)
			r.localCache.SetDefault(key, channel)
			break
		}
	}
	return
}

func (r *RoomKeysType) GetKey(channel string) (newKey string, err error) {

	var key interface{}
	var found bool
	if key, found = r.localCache.Get(channel); found {
		return key.(string), nil
	}
	newKey, err = r.SetKey(channel)
	logger.Info("New channel [%s]: %s", channel, newKey)
	return
}

func (r *RoomKeysType) GetChannel(key string) (channel string, err error) {

	chann, found := r.localCache.Get(key)
	if found {
		return chann.(string), nil
	} else {
		return "", fmt.Errorf("%s does not exists", key)
	}
}

func (r *RoomKeysType) DeleteChannel(channel string) bool {

	key, ok := r.localCache.Get(channel)
	if ok {
		r.localCache.Delete(channel)
		r.localCache.Delete(key.(string))
		return true
	}
	return false
}

func (r *RoomKeysType) DeleteKey(key string) bool {

	channel, ok := r.localCache.Get(key)
	if ok {
		r.localCache.Delete(channel.(string))
		r.localCache.Delete(key)
		return true
	}
	return false
}
