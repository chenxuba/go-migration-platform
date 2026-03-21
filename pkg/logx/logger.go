package logx

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Entry map[string]any

var std = log.New(os.Stdout, "", 0)

func Info(message string, fields Entry) {
	write("INFO", message, fields)
}

func Error(message string, fields Entry) {
	write("ERROR", message, fields)
}

func write(level, message string, fields Entry) {
	if fields == nil {
		fields = Entry{}
	}

	fields["level"] = level
	fields["message"] = message
	fields["time"] = time.Now().Format(time.RFC3339)

	payload, err := json.Marshal(fields)
	if err != nil {
		std.Printf(`{"level":"ERROR","message":"marshal log failed","time":"%s"}`+"\n", time.Now().Format(time.RFC3339))
		return
	}

	std.Println(string(payload))
}
