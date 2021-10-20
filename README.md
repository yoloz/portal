# PORTAL

a portal for myself, like note, reading, etc.

# Editormd

use [Editormd](https://pandao.github.io/editor.md/) write new page

# source

```

.
├── go.mod
├── go.sum
├── portal.go
├── portal_test.go
├── post
│   ├── docsify.go
│   └── upload.go
├── public
│   ├── bin
│   │   ├── start.sh
│   │   └── stop.sh
│   ├── config
│   │   ├── cert.pem
│   │   ├── config.json
│   │   └── key.pem
│   └── static
│       ├── editormd
│       ├── index.html
│       └── newPage.html
└── README.md

```

# runtime

```
.
├── bin
│   ├── portal
│   ├── start.sh
│   └── stop.sh
├── config
│   ├── cert.pem
│   ├── config.json
│   └── key.pem
└── static
    ├── editormd
    ├── index.html
    └── newPage.html
```

# classify

edit `static/index.html` and add

```html
  <div class="card">
            <a href="/alias" target="_blank">classify name</a>
  </div>
```

