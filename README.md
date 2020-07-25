# High Heath Farm Cattery

_version: 3_

This is the Hugo static site for www.highheathcattery.co.uk. Forms are sent to
a small golang server, and emails sent with Gmail API. Emails are created with
[hermes](https://github.com/matcornic/hermes/) and form data parsed with
[schemas](http://www.gorillatoolkit.org/pkg/schema).

## Getting Started

There is a docker compose file which will start the go app which runs
the static file server and form actions.

For fast theme development run:

    hugo server

To test forms you will need:

    docker-compose up


## Theme layout

### Pages

Normal pages are in `content/<page>.md`. The theme for this page will be looked
up following the usual look up order. In this project that is usually:

* `theme/highheath/layouts/_default/single.html`
* `theme/highheath/layouts/<page>/list.html`  # if it is a content type
