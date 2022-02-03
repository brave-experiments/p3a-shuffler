#!/usr/bin/env python3

import re
import os
from os.path import join, isfile
import sys
import argparse
import requests

ENCLAVE_ENDPOINT = ""


class S3FileReader(object):
    """
    Reads P3A measurements (as they are stored in the S3 bucket) from disk and
    replays them to our live Nitro enclave.  This object must act as an
    iterator.
    """
    def __init__(self, directory):
        self.dir = directory
        self.files = []
        self.cache = []

    def __iter__(self):
        self.files = [join(self.dir, f) for f in os.listdir(self.dir) if
                      isfile(join(self.dir, f))]
        self.files = sorted(self.files, reverse=True)
        return self

    def __next__(self):
        if len(self.cache) > 0:
            return self.cache.pop(0)

        try:
            file_content = self._read_file(self.files.pop())
        except IndexError:
            raise StopIteration

        measurements = self._extract_p3a_measurement(file_content)
        if len(measurements) > 1:
            self.cache += measurements
            return self.cache.pop(0)
        else:
            return measurements[0]

    def _read_file(self, filename):
        with open(filename, "r") as fd:
            return fd.readlines()

    def _extract_p3a_measurement(self, file_content):
        measurements = []
        # There may be multiple measurements per file.
        for line in file_content:
            m = re.search("'([^']+)'", file_content[0])
            if m:
                # print(m.group(1))
                measurements.append(m.group(1))
        return measurements


def replay_p3a_data(directory):
    measurements = S3FileReader(directory)
    for m in measurements:
        # Add "verify=False" if the enclave is running in testing mode, and
        # using self-signed certificates.
        requests.post(ENCLAVE_ENDPOINT, data="[%s]" % m)


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: %s DIRECTORY" % sys.argv[0], file=sys.stderr)
        sys.exit(1)
    replay_p3a_data(sys.argv[1])
