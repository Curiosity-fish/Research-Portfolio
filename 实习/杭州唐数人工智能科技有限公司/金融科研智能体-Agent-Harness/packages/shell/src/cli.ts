#!/usr/bin/env node

import { shellMain } from "./entry.ts";

const exitCode = await shellMain(process.argv.slice(2));
process.exitCode = exitCode;
