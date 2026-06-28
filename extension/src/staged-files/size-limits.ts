/** Max raw file bytes staged locally (matches server draft upload default). */
export const MAX_STAGED_FILE_BYTES = 500 * 1024 * 1024;

export function isStagedFileTooLarge(byteLength: number): boolean {
  return byteLength > MAX_STAGED_FILE_BYTES;
}

export function stagedFileSizeError(fileName: string, byteLength: number): string {
  const mb = (byteLength / (1024 * 1024)).toFixed(1);
  const maxMb = (MAX_STAGED_FILE_BYTES / (1024 * 1024)).toFixed(0);
  return `${fileName} (${mb} MB) exceeds Rubrical's ${maxMb} MB per-file limit`;
}
