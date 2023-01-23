package main

import (
	"C"

	"fmt"
	"log"
	"strings"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
)

var (
	msgKey       = "message"
	tsLayout     = "20060102T15:04:05Z"
	tsLoc        *time.Location
	optKeys      []string
	lastMsgByTag = map[string]string{}
	skipDupMsg   bool
	floorFloat   bool
)

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	return output.FLBPluginRegister(def, "telegram", "Telegram Output Plugin.")
}

// (fluentbit will call this)
// plugin (context) pointer to fluentbit context (state/ c code)
//
//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	getParam := func(key string) string {
		return output.FLBPluginConfigKey(plugin, key)
	}

	tgApiToken := getParam("api_token")
	tgRoomIDs := getParam("room_ids")
	if err := initTgBot(tgApiToken, tgRoomIDs); err != nil {
		log.Printf("fail to init telegram bot: %v", err)
		return output.FLB_ERROR
	}

	if getParam("message_key") != "" {
		msgKey = getParam("message_key")
	}

	if getParam("timestamp_layout") != "" {
		tsLayout = getParam("timestamp_layout")
	}

	if getParam("timestamp_location") != "" {
		var err error
		tsLoc, err = time.LoadLocation(getParam("timestamp_location"))
		if err != nil {
			log.Printf("fail to load location: %v", err)
			return output.FLB_ERROR
		}
	} else {
		tsLoc, _ = time.LoadLocation("UTC")
	}

	if getParam("optional_keys") != "" {
		optKeys = strings.Split(getParam("optional_keys"), ",")
		for i, v := range optKeys {
			optKeys[i] = strings.TrimSpace(v)
		}
	}

	if getParam("surpress_duplication") == "yes" {
		skipDupMsg = true
	}

	if getParam("floor_float") == "yes" {
		floorFloat = true
	}

	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	var ret int
	var ts interface{}
	var record map[interface{}]interface{}
	dec := output.NewDecoder(data, int(length)) // Create Fluent Bit decoder

	// count := 0 // batch out count
	for {
		ret, ts, record = output.GetRecord(dec)
		if ret != 0 { // all record have been flushed
			break
		}

		valueByKey := map[string]string{}
		for k, v := range record {
			valueByKey[str(k)] = str(v)
		}

		var msg string
		var ok bool
		if msg, ok = valueByKey[msgKey]; !ok {
			log.Printf("message key not found: %v", msgKey)
			return output.FLB_ERROR
		}

		if lastMsg, ok := lastMsgByTag[str(tag)]; ok && skipDupMsg && lastMsg == msg {
			continue
		}
		lastMsgByTag[str(tag)] = msg

		tsStr := getTime(ts).In(tsLoc).Format(tsLayout)
		var optMsg string
		for _, k := range optKeys {
			if v, ok := valueByKey[k]; ok {
				optMsg += fmt.Sprintf("- %s: %s\n", k, v)
			}
		}
		if optMsg != "" {
			msg = fmt.Sprintf(
				"%s\n---\n%s---\n%s",
				msg, optMsg, tsStr,
			)
		} else {
			msg = fmt.Sprintf(
				"%s\n---\n%s",
				msg, tsStr,
			)
		}

		if err := sendMsgToTelegram(msg); err != nil {
			log.Printf("fail to send msg to telegram: %v", err)
			return output.FLB_ERROR
		}
		// count++
	}

	// Return options:
	//
	// output.FLB_OK    = data have been processed.
	// output.FLB_ERROR = unrecoverable error, do not try this again.
	// output.FLB_RETRY = retry to flush later.
	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

// ---

func str(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case *C.char:
		return C.GoString(v)
	case float64:
		if floorFloat {
			return fmt.Sprintf("%d", int(v+0.5))
		} else {
			return fmt.Sprintf("%f", v)
		}
	case float32:
		if floorFloat {
			return fmt.Sprintf("%d", int(v+0.5))
		} else {
			return fmt.Sprintf("%f", v)
		}
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getTime(ts any) time.Time {
	var timestamp time.Time
	switch t := ts.(type) {
	case output.FLBTime:
		timestamp = ts.(output.FLBTime).Time
	case uint64:
		timestamp = time.Unix(int64(t), 0)
	default:
		fmt.Println("time provided invalid, defaulting to now.")
		timestamp = time.Now()
	}
	return timestamp
}

func main() {
}
