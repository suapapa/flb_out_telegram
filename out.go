package main

import (
	"C"

	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
)

// pluginState holds runtime configuration and the Telegram sender for the Fluent Bit plugin.
type pluginState struct {
	telegram *TelegramOutput

	msgKey       string
	tsLayout     string
	tsLoc        *time.Location
	optKeys      []string
	lastMsgByTag map[string]string
	lastMsgTs    map[string]time.Time

	skipDupMsg    bool
	skipDupMsgDur time.Duration
	floorFloat    bool
}

var state pluginState

var pluginVersion = "dev"

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	return output.FLBPluginRegister(def, "telegram", "Telegram Output Plugin.")
}

// (fluentbit will call this)
// plugin (context) pointer to fluentbit context (state/ c code)
//
//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	if state.telegram != nil {
		log.Printf("telegram plugin already initialized")
		return output.FLB_ERROR
	}

	getParam := func(key string) string {
		p := output.FLBPluginConfigKey(plugin, key)

		// Trim inline comment
		if i := strings.Index(p, "#"); i != -1 {
			p = p[:i]
		}

		return strings.TrimSpace(p)
	}

	isParamTrue := func(key string) bool {
		switch p := strings.ToLower(getParam(key)); p {
		case "yes", "on", "true":
			return true
		}
		return false
	}

	tgAPI := getParam("api_token")
	tgRooms := getParam("room_ids")
	tgOut, err := NewTelegramOutput(tgAPI, tgRooms)
	if err != nil {
		log.Printf("fail to init telegram bot: %v", err)
		return output.FLB_ERROR
	}

	st := pluginState{
		telegram:     tgOut,
		msgKey:       "message",
		tsLayout:     "20060102T15:04:05Z",
		lastMsgByTag: map[string]string{},
		lastMsgTs:    map[string]time.Time{},
	}

	if getParam("message_key") != "" {
		st.msgKey = getParam("message_key")
	}

	if getParam("timestamp_layout") != "" {
		st.tsLayout = getParam("timestamp_layout")
	}

	if getParam("timestamp_location") != "" {
		var locErr error
		st.tsLoc, locErr = time.LoadLocation(getParam("timestamp_location"))
		if locErr != nil {
			log.Printf("fail to load location: %v", locErr)
			return output.FLB_ERROR
		}
	} else {
		st.tsLoc, _ = time.LoadLocation("UTC")
	}

	if getParam("optional_keys") != "" {
		st.optKeys = strings.Split(getParam("optional_keys"), ",")
		for i, v := range st.optKeys {
			st.optKeys[i] = strings.TrimSpace(v)
		}
	}

	st.skipDupMsg = isParamTrue("suppress_duplication")
	st.floorFloat = isParamTrue("floor_float")

	st.skipDupMsgDur, err = time.ParseDuration(getParam("suppress_timeout"))
	if err != nil {
		log.Printf("fail to parse suppress_timeout: %v", err)
		// not fatal
	}

	state = st

	fmt.Printf("telegram output plugin %s initialized\n", pluginVersion)
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	ctx := context.Background()

	var ret int
	var ts interface{}
	var record map[interface{}]interface{}
	dec := output.NewDecoder(data, int(length)) // Create Fluent Bit decoder

	for {
		ret, ts, record = output.GetRecord(dec)
		if ret != 0 { // all record have been flushed
			break
		}

		valueByKey := make(map[string]string, len(record))
		for k, v := range record {
			valueByKey[formatKey(k)] = formatRecordValue(state.floorFloat, v)
		}

		msg, ok := valueByKey[state.msgKey]
		if !ok {
			log.Printf("message key not found: %v", state.msgKey)
			return output.FLB_ERROR
		}

		tagStr := formatRecordValue(false, tag)
		if lastMsg, dup := state.lastMsgByTag[tagStr]; dup && state.skipDupMsg {
			if lastMsg == msg && time.Since(state.lastMsgTs[tagStr]) < state.skipDupMsgDur {
				continue
			}
		}

		tsTime := parseRecordTime(ts)
		state.lastMsgByTag[tagStr] = msg
		state.lastMsgTs[tagStr] = tsTime

		tsStr := tsTime.In(state.tsLoc).Format(state.tsLayout)
		optMsg := buildOptionalLines(valueByKey, state.optKeys)
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

		if err := state.telegram.SendMessage(ctx, msg); err != nil {
			log.Printf("fail to send msg to telegram: %v", err)
			return output.FLB_ERROR
		}
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

func buildOptionalLines(valueByKey map[string]string, keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	var b strings.Builder
	for _, k := range keys {
		if v, ok := valueByKey[k]; ok {
			b.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
		}
	}
	return b.String()
}

func formatKey(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatRecordValue(floorFloat bool, v interface{}) string {
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
		}
		return fmt.Sprintf("%f", v)
	case float32:
		if floorFloat {
			return fmt.Sprintf("%d", int(v+0.5))
		}
		return fmt.Sprintf("%f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func parseRecordTime(ts any) time.Time {
	switch t := ts.(type) {
	case output.FLBTime:
		return t.Time
	case uint64:
		return time.Unix(int64(t), 0)
	default:
		log.Printf("invalid timestamp type %T, using now", ts)
		return time.Now()
	}
}

func main() {
}
