# Config [![GoDoc](https://godoc.org/github.com/jiyeyuran/go-config?status.svg)](https://godoc.org/github.com/jiyeyuran/go-config)

Go Config which is extracted from [go-micro](https://github.com/asim/go-micro/config) is a pluggable dynamic config package.

Most config in applications are statically configured or include complex logic to load from multiple sources. 
Go Config makes this easy, pluggable and mergeable. You'll never have to deal with config in the same way again.

## Features

- **Dynamic Loading** - Load configuration from multiple source as and when needed. Go Config manages watching config sources 
in the background and automatically merges and updates an in memory view. 

- **Pluggable Sources** - Choose from any number of sources to load and merge config. The backend source is abstracted away into 
a standard format consumed internally and decoded via encoders. Sources can be env vars, flags, file, etcd, k8s configmap, etc.

- **Mergeable Config** - If you specify multiple sources of config, regardless of format, they will be merged and presented in 
a single view. This massively simplifies priority order loading and changes based on environment.

- **Observe Changes** - Optionally watch the config for changes to specific values. Hot reload your app using Go Config's watcher. 
You don't have to handle ad-hoc hup reloading or whatever else, just keep reading the config and watch for changes if you need 
to be notified.

- **Sane Defaults** - In case config loads badly or is completely wiped away for some unknown reason, you can specify fallback 
values when accessing any config values directly. This ensures you'll always be reading some sane default in the event of a problem.

## Getting Started

Go Config has the benefit of supporting multiple backend sources and config encoding formats out of the box.

Here’s the top level interface which encapsulates all the features mentioned.
```go
// Config is an interface abstraction for dynamic configuration
type Config interface {
	// provide the reader.Values interface
	reader.Values
	// Init the config
	Init(opts ...Option) error
	// Options in the config
	Options() Options
	// Stop the config loader/watcher
	Close() error
	// Load config sources
	Load(source ...source.Source) error
	// Force a source changeset sync
	Sync() error
	// Watch a value for changes
	Watch(path ...string) (Watcher, error)
}
```
Ok so let’s break it down and discuss the various concerns in the framework starting with the backend sources.

### Source

A source is a backend from which config is loaded. This could be command line flags, environment variables, a key-value store or any other number of places.

Go Config provides a simple abstraction over all these sources as a simple interface from which we read data or what we call a ChangeSet.
```go
// Source is the source from which config is loaded
type Source interface {
	Read() (*ChangeSet, error)
	Write(*ChangeSet) error
	Watch() (Watcher, error)
	String() string
}

// ChangeSet represents a set of changes from a source
type ChangeSet struct {
	Data      []byte
	Checksum  string
	Format    string
	Source    string
	Timestamp time.Time
}
```
The ChangeSet includes the raw data, it’s format, timestamp of creation or last update and the source from which it was loaded. There’s also an optional md5 checksum which can be recalculated using the Sum() method.

The simplicity of this interface allows us to easily create a source for any backend, read it’s values at any given time or watch for changes where possible.

### Encoding

Config is rarely available in just a single format and people usually have varying preferences on whether it should be stored in json, yaml, toml or something else. We make sure to deal with this in the framework so almost any encoding format can be dealt with.

The encoder is a very simply interface for handling encoding and decoding different formats. Why wouldn’t we reuse existing libraries for this? We do beneath the covers but to ensure we could deal with encoding in an abstract way it made sense to define an interface for it.
```go
// Encoder handles encoding and decoding of a variety of config formats
type Encoder interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
	String() string
}
```
The current supported formats are json, yaml, toml and xml.

### Reader

Once we’ve loaded backend sources and developed a way to decode the variety of config formats we need some way of actually internally representing and reading it. For this we’ve created a reader.

The reader manages decoding and merging multiple changesets into a single source of truth. It then provides a value interface which allows you to retrieve native Go types or scan the config into a type of your choosing.

```go
// Reader manages merging multiple changesets into a single source of truth
type Reader interface {
    Merge(...*source.ChangeSet) (*source.ChangeSet, error)
    Values(*source.ChangeSet) (Values, error)
    String() string
}

// Values is returned by the reader
type Values interface {
	Bytes() []byte
	Get(path ...string) Value
	Set(val interface{}, path ...string)
	Del(path ...string)
	Map() map[string]interface{}
	Scan(v interface{}) error
}

// Value represents a value of any type
type Value interface {
	Bool(def bool) bool
	Int(def int) int
	String(def string) string
	Float64(def float64) float64
	Duration(def time.Duration) time.Duration
	StringSlice(def []string) []string
	StringMap(def map[string]string) map[string]string
	Scan(val interface{}) error
	Bytes() []byte
}
```
Our default internal representation for the merged source is json.

### Example

Let’s look at how Go Config actually works in code. Starting with a simple example, let’s read config from a file.

### Read Config

Step 1. Define a config.json file
```json
{
    "hosts": {
        "database": {
            "address": "10.0.0.1",
            "port": 3306
        },
        "cache": {
            "address": "10.0.0.2",
            "port": 6379
        }
    }
}
```
Step 2. Load the file into config
```go
config.Load(file.NewSource(
	file.WithPath("config.json"),
))
```
Step 3. Read the values from config
```go
type Host struct {
	Address string `json:"address"`
	Port int `json:"port"`
}

var host Host

config.Get("hosts", "database").Scan(&host)
```
And that’s it! It’s really that simple.

### Watch Config

If the config file changes, the next time you read the value it will be different. But what if you want to track that change? You can watch for changes. Let’s test it out.
```go
w, err := config.Watch("hosts", "database")
if err != nil {
	// do something
}

// wait for next value
v, err := w.Next()
if err != nil {
	// do something
}

var host Host

v.Scan(&host)
```
In this example rather than getting a value, we watch it. The next time the value changes we’ll receive it and can update our Host struct.

### Merge Config

Another useful feature is the ability to load config from multiple sources which are ordered, merged and overridden. A good example of this would be loading config from a file but overriding via environment variables or flags.
```go
config.Load(
	// base config
	file.NewSource(),
	// override file with env vars
	env.NewSource(),
	// override env vars with flags
	flag.NewSource(),
)
```

### Fallback Values
```go
// Get address. Set default to localhost as fallback
address := config.Get("hosts", "database", "address").String("localhost")

// Get port. Set default to 3000 as fallback
port := config.Get("hosts", "database", "port").Int(3000)
```

### Summary

The way in which config is managed and consumed needs to evolve. Go Config looks to do this by drastically simplifying use of dynamic configuration with a pluggable framework.

Go Config currently supports a number of configuration formats and backend sources but we’re always looking for more contributions. If you’re interested in contribution please feel free to do so by with a pull request.

Let Go Config managed the complexity of configuration for you so you can focus on what’s really important. Your code.
