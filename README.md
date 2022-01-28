Shuffler
========

This repository implements a shuffler, similar to how it was originally
proposed in the
[PROCHLO](https://arxiv.org/pdf/1710.00901.pdf)
paper.  The shuffer is meant to run in an
[AWS Nitro Enclave](https://aws.amazon.com/ec2/nitro/nitro-enclaves/)
which allows third parties to verify the code's authenticity.  The shuffler
takes as input
[telemetry measurements from Brave clients](https://github.com/brave/brave-browser/wiki/P3A),
discards measurements that are not shared by at least k other clients, and
periodically forwards them to the backend.

Data flow
---------

The Web API receives incoming requests and forwards them to the shuffler.  The
shuffler uses a briefcase to store requests, remove requests that don't satisfy
our k-anonymity protections, and periodically hands them over to the forwarder,
which forwards remaining requests to our backend.

    ┌────────┐  ┌──────────┐  ┌──────────┐
    │ WebAPI │─▶│ Shuffler │─▶│ Forwarder│
    └────────┘  └──────────┘  └──────────┘
                     │
                     ▼
               ┌───────────┐
               │ Briefcase │
               └───────────┘

Input
-----

The shuffler exposes the following HTTP endpoint:

    POST <endpoint>/reports

The request body contains a JSON-formatted list of P3A measurements, e.g.:

    [
      {
        "channel":"developer",
        "country_code":"US",
        "metric_name":"Brave.Welcome.InteractionStatus",
        "metric_value":3,
        "platform":"linux-bc",
        "refcode":"none",
        "version":"1.36.68",
        "woi":4,
        "wos":4,
        "yoi":2022,
        "yos":2022
      },
      ...
    ]

Output
------

The shuffler discards measurements that don't satisfy our k-anonymity
thresholds and forwards remaining measurements periodically to its backend.
