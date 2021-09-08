# High Heath Farm Cattery

_version: 3_ [![Coverage Status](https://coveralls.io/repos/github/Smirl/highheath/badge.svg?branch=master)](https://coveralls.io/github/Smirl/highheath?branch=master)

This is the Hugo static site for www.highheathcattery.co.uk. Forms are sent to
a small golang server, and emails sent with Gmail API. Emails are created with
[hermes](https://github.com/matcornic/hermes/) and form data parsed with
[schemas](http://www.gorillatoolkit.org/pkg/schema).

Comments automatically create pull requests and send confirmation emails. This
is inspired by staticman. This repo is installed as a Github App with permission
to edit content and pull requests.

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

### New Relic

There is the `<script>` tag in the `<head>` for New Relic Browser integration.
The snippet is available in the New Relic UI.


## Deployment

Github Actions are triggered to build and deploy the app on release. Release
Drafter is used to draft releases based on pull request titles. A service
account is used to deploy via helm3. This must be created first.

### Cleaning old docker images

A helper to clean docker images from the registry after many tags have been pushed.

`python docker/clean_images.py` for a dry run. Add `--delete` to actually delete
them.

### Github Actions Setup

To create the service account and permissions, a cluster-admin needs to apply
the following:

```console
kubectl apply -f deploy/serviceaccount.yaml
```

[Secrets][github-actions-secrets] for github actions are as follows:

- `DOCKER_PASSWORD`: _The password for registry.smirlwebs.com_
- `K8S_SECRET`: _The full yaml secret for the serviceaccount_
- `K8S_URL` : _The url of the kubernetes api server_

Secrets in the cluster need to be created for `gmail` and `github`
authentication.

### Gmail API Setup

For the gmail api, a project [_Website Form Backend_][gmail-console] has been
created with Oauth2 credentials. `credentials.json` can be downloaded from the
console. The application should be started where, after authenticating the app,
a `token.json` will be created. These should not be checked in. A secret in the
cluster should be created:

```console
kubectl create secret generic gmail -n highheath --from-file=credentials.json --from-file=token.json
```

### Github API Setup

A Github App (not Oauth2 App) has been created called
[high-heath-farm-cattery][github-console]. This is installed for this repo only.
A private key has been created (and can be rotated) and should be downloaded
as `private-key.pem`. This should not be checked in. A secret in the
cluster should be created:

```console
kubectl create secret generic github -n highheath --from-file=private-key.pem
```

### Recaptcha Setup

Recaptcha V3 is used to protect the site. The users browser makes a request
before each form and a token is added to the form. This token is then checked
against the recaptcha API with a client secret. This secret needs to be created
with:

```console
kubectl create secret generic recaptcha --from-literal=secret=<SECRET>
```

The [recaptcha admin console][recaptcha] can be used to see the number of
requests and suspicious requests to the forms.

### Forestry.io Setup

Forestry.io is available at https://www.highheathcattery.co.uk/admin as a CMS.
The configuration is managed via that admin pannel and is found in the `.forestry`
directory at the root of the repo.

[github-actions-secrets]: https://github.com/Smirl/highheath/settings/secrets
[gmail-console]: https://console.developers.google.com/apis/dashboard?project=website-form-bac-1595189229489&folder=&organizationId=
[github-console]: https://github.com/settings/apps/high-heath-farm-cattery
[recaptcha]: https://www.google.com/recaptcha/admin/site/432158062
