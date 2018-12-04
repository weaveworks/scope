#!/usr/bin/env python

# List all available versions of Weave Net's dependencies:
# - Go
# - Docker
# - Kubernetes
#
# Depending on the parameters passed, it can gather the equivalent of the below
# bash one-liners:
#   git ls-remote --tags https://github.com/golang/go \
#     | grep -oP '(?<=refs/tags/go)[\.\d]+$' \
#     | sort --version-sort
#   git ls-remote --tags https://github.com/golang/go \
#     | grep -oP '(?<=refs/tags/go)[\.\d]+rc\d+$' \
#     | sort --version-sort \
#     | tail -n 1
#   git ls-remote --tags https://github.com/docker/docker \
#     | grep -oP '(?<=refs/tags/v)\d+\.\d+\.\d+$' \
#     | sort --version-sort
#   git ls-remote --tags https://github.com/docker/docker \
#     | grep -oP '(?<=refs/tags/v)\d+\.\d+\.\d+\-rc\d*$' \
#     | sort --version-sort \
#     | tail -n 1
#   git ls-remote --tags https://github.com/kubernetes/kubernetes \
#     | grep -oP '(?<=refs/tags/v)\d+\.\d+\.\d+$' \
#     | sort --version-sort
#   git ls-remote --tags https://github.com/kubernetes/kubernetes \
#     | grep -oP '(?<=refs/tags/v)\d+\.\d+\.\d+\-beta\.\d+$' \
#     | sort --version-sort | tail -n 1
#
# Dependencies:
# - python
# - git
#
# Testing:
# $ python -m doctest -v list_versions.py

from os import linesep, path
from sys import argv, exit, stdout, stderr
from getopt import getopt, GetoptError
from subprocess import Popen, PIPE
from pkg_resources import parse_version
from itertools import groupby
from six.moves import filter
import shlex
import re

# See also: /usr/include/sysexits.h
_ERROR_RUNTIME = 1
_ERROR_ILLEGAL_ARGS = 64

_TAG_REGEX = '^[0-9a-f]{40}\s+refs/tags/%s$'
_VERSION = 'version'
DEPS = {
    'go': {
        'url': 'https://github.com/golang/go',
        're': 'go(?P<%s>[\d\.]+(?:rc\d)*)' % _VERSION,
        'min': None
    },
    'docker': {
        'url': 'https://github.com/docker/docker',
        're': 'v(?P<%s>\d+\.\d+\.\d+(?:\-rc\d)*)' % _VERSION,
        # Weave Net only works with Docker from 1.10.0 onwards, so we ignore
        # all previous versions:
        'min': '1.10.0',
    },
    'kubernetes': {
        'url': 'https://github.com/kubernetes/kubernetes',
        're': 'v(?P<%s>\d+\.\d+\.\d+(?:\-beta\.\d)*)' % _VERSION,
        # Weave Kube requires Kubernetes 1.4.2+, so we ignore all previous
        # versions:
        'min': '1.4.2',
    }
}


class Version(object):
    ''' Helper class to parse and manipulate (sort, filter, group) software
    versions. '''

    def __init__(self, version):
        self.version = version
        self.digits = [
            int(x) if x else 0
            for x in re.match('(\d*)\.?(\d*)\.?(\d*).*?', version).groups()
        ]
        self.major, self.minor, self.patch = self.digits
        self.__parsed = parse_version(version)
        self.is_rc = self.__parsed.is_prerelease

    def __lt__(self, other):
        return self.__parsed.__lt__(other.__parsed)

    def __gt__(self, other):
        return self.__parsed.__gt__(other.__parsed)

    def __le__(self, other):
        return self.__parsed.__le__(other.__parsed)

    def __ge__(self, other):
        return self.__parsed.__ge__(other.__parsed)

    def __eq__(self, other):
        return self.__parsed.__eq__(other.__parsed)

    def __ne__(self, other):
        return self.__parsed.__ne__(other.__parsed)

    def __str__(self):
        return self.version

    def __repr__(self):
        return self.version


def _read_go_version_from_dockerfile():
    # Read Go version from weave/build/Dockerfile
    dockerfile_path = path.join(
        path.dirname(path.dirname(path.dirname(path.realpath(__file__)))),
        'build', 'Dockerfile')
    with open(dockerfile_path, 'r') as f:
        for line in f:
            m = re.match('^FROM golang:(\S*)$', line)
            if m:
                return m.group(1)
    raise RuntimeError(
        "Failed to read Go version from weave/build/Dockerfile."
        " You may be running this script from somewhere else than weave/tools."
    )


def _try_set_min_go_version():
    ''' Set the current version of Go used to build Weave Net's containers as
    the minimum version. '''
    try:
        DEPS['go']['min'] = _read_go_version_from_dockerfile()
    except IOError as e:
        stderr.write('WARNING: No minimum Go version set. Root cause: %s%s' %
                     (e, linesep))


def _sanitize(out):
    return out.decode('ascii').strip().split(linesep)


def _parse_tag(tag, version_pattern, debug=False):
    ''' Parse Git tag output's line using the provided `version_pattern`, e.g.:
    >>> _parse_tag(
        '915b77eb4efd68916427caf8c7f0b53218c5ea4a    refs/tags/v1.4.6',
        'v(?P<version>\d+\.\d+\.\d+(?:\-beta\.\d)*)')
    '1.4.6'
    '''
    pattern = _TAG_REGEX % version_pattern
    m = re.match(pattern, tag)
    if m:
        return m.group(_VERSION)
    elif debug:
        stderr.write(
            'ERROR: Failed to parse version out of tag [%s] using [%s].%s' %
            (tag, pattern, linesep))


def get_versions_from(git_repo_url, version_pattern):
    ''' Get release and release candidates' versions from the provided Git
    repository. '''
    git = Popen(
        shlex.split('git ls-remote --tags %s' % git_repo_url), stdout=PIPE)
    out, err = git.communicate()
    status_code = git.returncode
    if status_code != 0:
        raise RuntimeError('Failed to retrieve git tags from %s. '
                           'Status code: %s. Output: %s. Error: %s' %
                           (git_repo_url, status_code, out, err))
    return list(
        filter(None, (_parse_tag(line, version_pattern)
                      for line in _sanitize(out))))


def _tree(versions, level=0):
    ''' Group versions by major, minor and patch version digits. '''
    if not versions or level >= len(versions[0].digits):
        return  # Empty versions or no more digits to group by.
    versions_tree = []
    for _, versions_group in groupby(versions, lambda v: v.digits[level]):
        subtree = _tree(list(versions_group), level + 1)
        if subtree:
            versions_tree.append(subtree)
    # Return the current subtree if non-empty, or the list of "leaf" versions:
    return versions_tree if versions_tree else versions


def _is_iterable(obj):
    '''
    Check if the provided object is an iterable collection, i.e. not a string,
    e.g. a list, a generator:
    >>> _is_iterable('string')
    False
    >>> _is_iterable([1, 2, 3])
    True
    >>> _is_iterable((x for x in [1, 2, 3]))
    True
    '''
    return hasattr(obj, '__iter__') and not isinstance(obj, str)


def _leaf_versions(tree, rc):
    '''
    Recursively traverse the versions tree in a depth-first fashion,
    and collect the last node of each branch, i.e. leaf versions.
    '''
    versions = []
    if _is_iterable(tree):
        for subtree in tree:
            versions.extend(_leaf_versions(subtree, rc))
        if not versions:
            if rc:
                last_rc = next(filter(lambda v: v.is_rc, reversed(tree)), None)
                last_prod = next(
                    filter(lambda v: not v.is_rc, reversed(tree)), None)
                if last_rc and last_prod and (last_prod < last_rc):
                    versions.extend([last_prod, last_rc])
                elif not last_prod:
                    versions.append(last_rc)
                else:
                    # Either there is no RC, or we ignore the RC as older than
                    # the latest production version:
                    versions.append(last_prod)
            else:
                versions.append(tree[-1])
    return versions


def filter_versions(versions, min_version=None, rc=False, latest=False):
    ''' Filter provided versions

    >>> filter_versions(
        ['1.0.0-beta.1', '1.0.0', '1.0.1', '1.1.1', '1.1.2-rc1', '2.0.0'],
        min_version=None,    latest=False, rc=False)
    [1.0.0, 1.0.1, 1.1.1, 2.0.0]

    >>> filter_versions(
        ['1.0.0-beta.1', '1.0.0', '1.0.1', '1.1.1', '1.1.2-rc1', '2.0.0'],
        min_version=None,    latest=True,  rc=False)
    [1.0.1, 1.1.1, 2.0.0]

    >>> filter_versions(
        ['1.0.0-beta.1', '1.0.0', '1.0.1', '1.1.1', '1.1.2-rc1', '2.0.0'],
        min_version=None,    latest=False, rc=True)
    [1.0.0-beta.1, 1.0.0, 1.0.1, 1.1.1, 1.1.2-rc1, 2.0.0]

    >>> filter_versions(
        ['1.0.0-beta.1', '1.0.0', '1.0.1', '1.1.1', '1.1.2-rc1', '2.0.0'],
        min_version='1.1.0', latest=False, rc=True)
    [1.1.1, 1.1.2-rc1, 2.0.0]

    >>> filter_versions(
        ['1.0.0-beta.1', '1.0.0', '1.0.1', '1.1.1', '1.1.2-rc1', '2.0.0'],
        min_version=None,    latest=True,  rc=True)
    [1.0.1, 1.1.1, 1.1.2-rc1, 2.0.0]

    >>> filter_versions(
        ['1.0.0-beta.1', '1.0.0', '1.0.1', '1.1.1', '1.1.2-rc1', '2.0.0'],
        min_version='1.1.0', latest=True,  rc=True)
    [1.1.1, 1.1.2-rc1, 2.0.0]
    '''
    versions = sorted([Version(v) for v in versions])
    if min_version:
        min_version = Version(min_version)
        versions = [v for v in versions if v >= min_version]
    if not rc:
        versions = [v for v in versions if not v.is_rc]
    if latest:
        versions_tree = _tree(versions)
        return _leaf_versions(versions_tree, rc)
    else:
        return versions


def _usage(error_message=None):
    if error_message:
        stderr.write('ERROR: ' + error_message + linesep)
    stdout.write(
        linesep.join([
            'Usage:', '    list_versions.py [OPTION]... [DEPENDENCY]',
            'Examples:', '    list_versions.py go',
            '    list_versions.py -r docker',
            '    list_versions.py --rc docker',
            '    list_versions.py -l kubernetes',
            '    list_versions.py --latest kubernetes', 'Options:',
            '-l/--latest Include only the latest version of each major and'
            ' minor versions sub-tree.',
            '-r/--rc     Include release candidate versions.',
            '-h/--help   Prints this!', ''
        ]))


def _validate_input(argv):
    try:
        config = {'rc': False, 'latest': False}
        opts, args = getopt(argv, 'hlr', ['help', 'latest', 'rc'])
        for opt, value in opts:
            if opt in ('-h', '--help'):
                _usage()
                exit()
            if opt in ('-l', '--latest'):
                config['latest'] = True
            if opt in ('-r', '--rc'):
                config['rc'] = True
        if len(args) != 1:
            raise ValueError('Please provide a dependency to get versions of.'
                             ' Expected 1 argument but got %s: %s.' %
                             (len(args), args))
        dependency = args[0].lower()
        if dependency not in DEPS.keys():
            raise ValueError(
                'Please provide a valid dependency.'
                ' Supported one dependency among {%s} but got: %s.' %
                (', '.join(DEPS.keys()), dependency))
        return dependency, config
    except GetoptError as e:
        _usage(str(e))
        exit(_ERROR_ILLEGAL_ARGS)
    except ValueError as e:
        _usage(str(e))
        exit(_ERROR_ILLEGAL_ARGS)


def main(argv):
    try:
        dependency, config = _validate_input(argv)
        if dependency == 'go':
            _try_set_min_go_version()
        versions = get_versions_from(DEPS[dependency]['url'],
                                     DEPS[dependency]['re'])
        versions = filter_versions(versions, DEPS[dependency]['min'], **config)
        print(linesep.join(map(str, versions)))
    except Exception as e:
        print(str(e))
        exit(_ERROR_RUNTIME)


if __name__ == '__main__':
    main(argv[1:])
