# How to generate these screenshots

* We use [msc-gen](https://gitlab.com/msc-generator/msc-generator) to
  draw the signalling chart from textual description.
* Screenshots can be taken with a program such as
  [flameshot](https://flameshot.org/), or whatever program your
  desktop environment provides by default.

1. Spin-up Shoelaces

You can do this from the root directory of this git repository:
```
make
./shoelaces -config configs/shoelaces.conf
```

2. Generate `shoelaces-frontend-1.png` by visiting
   [localhost:8081](http://localhost:8081) with a graphical web
   browser. Then, use your screenshot program.
   
3. Generate `shoelaces-frontend-2.png` by making a few requests with
   fake MAC addresses. You can do this from a new tab on your web
   browser, or using a CLI program such as `curl`. For example:
```
curl http://localhost:8081/poll/1/06-66-de-ad-be-ef
curl http://localhost:8081/poll/1/06-66-d3-ad-b3-3f
```
   Then, select one of thes servers, then a boot targets such as
   `debian.ipxe`, and use your screenshot program.
   
4. Generate `shoelaces-overview.png` by simply running `make` in the
   `/doc/screenshots/` directory. You will need `msc-gen`.
