# Caching

The `App` struct provides a simple thread-safe cache for storing and retrieving arbitrary data.

### Adding to the Cache

```go
app.Cache.Put("key", "value")
```

### Retrieving from the Cache

```go
value, err := app.Cache.Get("key")
```

### Lazy Initialization

Objects can be lazily initialized by providing a function that returns the object.

```go
app.Cache.Put("key", func() any {
  // instantiate an object that is expensive to create
})

// later - the object is initialized only when it is first retrieved
value, err := app.Cache.Get("key")
```
