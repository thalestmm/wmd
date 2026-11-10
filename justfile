_help:
  @echo ""
  @just --list

g: 
  @templ generate

p msg="update" mode="chore": g
  @git add .
  @git commit -m "{{ mode }}: {{ msg }}"
  @git push

alias i := install

install:
  @go install github.com/thalestmm/wmd@latest
