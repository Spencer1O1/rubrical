export type Accessibility = "staged" | "saved" | "inaccessible" | "staging_failed";

export type StagedFileRecord = {
  assignmentKey: string;
  canvasFileId?: string;
  fileName: string;
  normalizedFileName: string;
  stagedAt: string;
  mimeType: string;
};

export type DraftManifestFile = {
  serverFileId: number;
  fileName: string;
  canvasFileId?: string;
  byteSize: number;
  uploadedAt: string;
};

export type DraftManifest = {
  assignmentId?: number;
  files: DraftManifestFile[];
};

export type RowAccessibility = {
  fileName: string;
  fileId: string | null;
  state: Accessibility;
  serverFileId?: number;
  stagedRecord?: Pick<StagedFileRecord, "normalizedFileName" | "stagedAt" | "canvasFileId">;
};

export type CanvasIdAssignment = {
  rowIndex: number;
  normalizedFileName: string;
  fileId: string;
};

export type ReconcilePromotion = {
  normalizedFileName: string;
  stagedAt: string;
  canvasFileId: string;
};
