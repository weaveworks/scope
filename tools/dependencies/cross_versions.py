#!/usr/bin/env python

# Generate the cross product of latest versions of Weave Net's dependencies:
# - Go
# - Docker
# - Kubernetes
#
# Dependencies:
# - python
# - git
# - list_versions.py
#
# Testing:
# $ python -m doctest -v cross_versions.py

from os import linesep
from sys import argv, exit, stdout, stderr
from getopt import getopt, GetoptError
from list_versions import DEPS, get_versions_from, filter_versions
from itertools import product

# See also: /usr/include/sysexits.h
_ERROR_RUNTIME = 1
_ERROR_ILLEGAL_ARGS = 64


def _usage(error_message=None):
    if error_message:
        stderr.write('ERROR: ' + error_message + linesep)
    stdout.write(
        linesep.join([
            'Usage:', '    cross_versions.py [OPTION]...', 'Examples:',
            '    cross_versions.py', '    cross_versions.py -r',
            '    cross_versions.py --rc', '    cross_versions.py -l',
            '    cross_versions.py --latest', 'Options:',
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
        if len(args) != 0:
            raise ValueError('Unsupported argument(s): %s.' % args)
        return config
    except GetoptError as e:
        _usage(str(e))
        exit(_ERROR_ILLEGAL_ARGS)
    except ValueError as e:
        _usage(str(e))
        exit(_ERROR_ILLEGAL_ARGS)


def _versions(dependency, config):
    return map(str,
               filter_versions(
                   get_versions_from(DEPS[dependency]['url'],
                                     DEPS[dependency]['re']),
                   DEPS[dependency]['min'], **config))


def cross_versions(config):
    docker_versions = _versions('docker', config)
    k8s_versions = _versions('kubernetes', config)
    return product(docker_versions, k8s_versions)


def main(argv):
    try:
        config = _validate_input(argv)
        print(linesep.join('\t'.join(triple)
                           for triple in cross_versions(config)))
    except Exception as e:
        print(str(e))
        exit(_ERROR_RUNTIME)


if __name__ == '__main__':
    main(argv[1:])
