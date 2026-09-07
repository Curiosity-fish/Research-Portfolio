import { type ClientCommandContext, clientCommand } from "./commands/client.ts";
import { type ServerCommandContext, serverCommand } from "./commands/server.ts";
import { type TfaCommandContext, tfaCommand } from "./commands/tfa.ts";

export type ExperimentalCliContext = TfaCommandContext & ServerCommandContext & ClientCommandContext;

export const experimentalCli = tfaCommand.command(serverCommand).command(clientCommand);
