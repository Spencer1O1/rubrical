export type RubricalFetchRequest = {
  path: string;
  method?: string;
  headers?: Record<string, string>;
  body?: string;
};

export type RubricalFetchSuccess = {
  ok: true;
  data: unknown;
  base: string;
};

export type RubricalFetchFailure = {
  ok: false;
  error: string;
};

export type RubricalFetchResult = RubricalFetchSuccess | RubricalFetchFailure;
