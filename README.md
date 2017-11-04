# mdloop

Simple livereload markdown preview server written in Go.

Useful for editing GitHub README files.

## Install
```
go get github.com/zaluska/mdloop
```

## Usage

1. Run:

    ```
    mdloop [-f filename.md]
    ```

    If no filename is specified then `README.md` is used as default.

2. Go to [localhost:8080](http://localhost:8080/) in your web browser.
