let strictExtraction = false;

export function isStrictExtraction(): boolean {
  return strictExtraction;
}

export function setStrictExtraction(value: boolean): void {
  strictExtraction = value;
}
