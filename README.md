# High Heath Farm Cattery

_version: 3_

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


## Deployment

Github Actions are triggered to build and deploy the app on release. Release
Drafter is used to draft releases based on pull request titles. A service
account is used to deploy via helm3. This must be created first.

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

[github-actions-secrets]: https://github.com/Smirl/highheath/settings/secrets
[gmail-console]: https://console.developers.google.com/apis/dashboard?project=website-form-bac-1595189229489&folder=&organizationId=
[github-console]: https://github.com/settings/apps/high-heath-farm-cattery