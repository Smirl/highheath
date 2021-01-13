"""Get Images."""

from argparse import ArgumentParser
from functools import total_ordering
import re
import requests
import os

URL = "https://registry.smirlwebs.com/v2"
USERNAME = "smirl"
PASSWORD = os.environ["PASSWORD"]
AUTH = (USERNAME, PASSWORD)
HEADERS = {"Accept": "application/vnd.docker.distribution.manifest.v2+json"}


@total_ordering
class SemVer:

    regex = re.compile("v?(?P<major>\d+)(?:\.(?P<minor>\d+)(?:\.(?P<patch>\d+))?)?")

    def __init__(self, major, minor=0, patch=0):
        self.major = int(major)
        self.minor = int(minor)
        self.patch = int(patch)

    def __eq__(self, other):
        return (
            self.major == other.major
            and self.minor == other.minor
            and self.patch == other.patch
        )

    def __lt__(self, other):
        return (
            self.major < other.major
            or (self.major == other.major and self.minor < other.minor)
            or (
                self.major == other.major
                and self.minor == other.minor
                and self.patch < other.patch
            )
        )

    def __str__(self):
        return f"SemVer(major={self.major}, minor={self.minor}, patch={self.patch})"

    @classmethod
    def parse(cls, value):
        if match := cls.regex.match(value):
            return cls(match.group("major"), match.group("minor"), match.group("patch"))
        else:
            raise TypeError("Cannot parse semver")


def get_repos():
    """Get all of the repos."""
    res = requests.get(f"{URL}/_catalog", auth=AUTH, headers=HEADERS)
    return res.json()["repositories"]


def get_tags(repo):
    """Get the tags for an repo."""
    res = requests.get(f"{URL}/{repo}/tags/list", auth=AUTH, headers=HEADERS)
    return res.json()["tags"]


def get_digest(repo, tag):
    """Get digest for image."""
    res = requests.get(f"{URL}/{repo}/manifests/{tag}", auth=AUTH, headers=HEADERS)
    return res.headers["Docker-Content-Digest"]


def delete_digest(repo, digest):
    """Delete the digest."""
    res = requests.delete(
        f"{URL}/{repo}/manifests/{digest}", auth=AUTH, headers=HEADERS
    )
    res.raise_for_status()


def main(delete=False):
    """List and optionally delete old docker images."""
    for repo in get_repos():
        print(repo)
        for tag in sorted(get_tags(repo), key=SemVer.parse)[:-3]:
            digest = get_digest(repo, tag)
            print("\tDELETE", "" if delete else "(dry run)", tag, digest)
            if delete is True:
                delete_digest(repo, digest)


if __name__ == "__main__":
    parser = ArgumentParser()
    parser.add_argument("--delete", action="store_true")
    args = parser.parse_args()
    main(args.delete)
