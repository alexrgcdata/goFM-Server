# 2. Types and configuration

Go groups data in structs. Config contains routes and log settings. JSON tags
such as `json:"routes"` tell encoding/json which property to read.

Slices such as []Route hold a variable number of routes. Maps such as
map[string]any hold normalized JSON data when fields are not known at compile
time.

Config.Validate is part of the security boundary. It rejects unsafe CORS
origins, invalid target URLs, duplicate method/path pairs, empty tokens, and
invalid log capacity before the server starts.
