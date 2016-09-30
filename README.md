# 🍊🍊🍊

Inspired by the [Heroku API Design Guide](https://www.gitbook.com/book/geemus/http-api-design/details) this package parses and generates `Range` headers. 

```
go get -u github.com/aj0strow/orange
```

### Example

Encode and decode ranges.

```go
r := orange.Range{
    Sort: "name",
    Start: "g",
    StartExclusive: true,
    End: "",
    Limit: 200,
    Desc: false,
}
r.String() == "name ~g..; max=200;"

r == orange.ParseString(r.String())
```

Get the next range by passing in the last value of the current result set.

```go
r := orange.ParseString("age ..; max=10,order=desc;")
r.Next("65").String() == "age ~65..; max=10,order=desc;"
```

Parse range header from the request.

```go
r, err := orange.From(req)
if err != nil {
    // invalid range header syntax
    http.Error(resp, err.Error(), 400)
    return
}
if r.Sort == "" {
    r.Sort = "default_property"
}
if r.Limit > 200 || r.Limit <= 0 {
    r.Limit = 200
}
```

Set range headers on the response.

```go
orange.Accept(resp, "id", "name")
orange.Next(resp, r.Next("65"))
```

### Request

Requests for list resources can specify the `Range` header with a sort property, range start and end, inclusive or exclusive start.

```sh
# Sort on name property.
Range: name;

# Sort on name property.
Range: name ..;

# Sort on name property starting at "a" end ending at "z".
Range: name a..z;

# Order descending this time.
Range: name z..a; order=desc;

# Exclusive start after "g".
Range: name ~g..z; order=asc;

# Exclusive start after "g" and limit to 50 records.
Range: name ~g..; max=50;
```

Some libraries like Microsoft C# restrict the range header to byte sequences. In that case use the `X-Range` header instead. 

### Response

Each response should include the `Accept-Range` header, which specifies which sort properties are supported.

```
Accept-Range: name, updated_at
```

If the entire range is returned, the response status should be `200 Ok`. For partial results the response status should be `206 Partial Content` (hey [Seattle](https://en.wikipedia.org/wiki/Area_code_206) 👋) and include a `Next-Range` header. 

```
Next-Range: name ~g..; max=50;
```

For example, let's load a list of friends alphabetically. You're really good looking and popular, so it's going to take several requests. 

```
GET /friends
Range: name ..; max=200;
```

```
206 Partial Content
Accept-Range: name, age
Next-Range: name ~Joey+Tribbiani..; max=200;
```

```
GET /friends
Range: name ~Joey+Tribbiani..; max=200;
```

```
200 Ok
Accept-Range: name, age
```

### License

MIT
