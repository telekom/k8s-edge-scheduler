# k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
# Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
# contact: opensource@telekom.de

# This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause]. 
# For Details see the file LICENSE on the top level of the project repository.

FROM scratch

COPY build/edge-scheduler /edge-scheduler

CMD [ "/edge-scheduler" ]
