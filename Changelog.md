### 1.6.7 - ENI 1.0.0

Features:

* Add ENI (Ethereum Native Interface) support.
  - Use CGO to execute native dynamic libraries.
  - Add ENI opcode.
  - Calculate ENI gas usage from dynamic libraries.
  - Parse ENI arguments and returned data to json format.

Bugfixes:

* Fix version comparison for go1.10 and beyond.
* Accept uppercase EVM instructions.
* Fix EVM hex number decoding problem.