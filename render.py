#!/usr/bin/env python3
from glob import glob
import json

with open("list.json", "w") as f:
    json.dump(glob("raw/*.json"), f)

