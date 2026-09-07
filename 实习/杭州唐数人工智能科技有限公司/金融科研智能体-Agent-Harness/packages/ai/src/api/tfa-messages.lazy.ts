import type { ProviderStreams } from "../types.ts";
import { lazyApi } from "./lazy.ts";

export const tfaMessagesApi = (): ProviderStreams => lazyApi(() => import("./tfa-messages.ts"));
