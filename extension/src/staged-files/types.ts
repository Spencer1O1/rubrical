export type Accessibility = "staged" | "saved" | "inaccessible";

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

export type StagedFilesMessage =
  | { type: "staged-files:ping" }
  | {
      type: "staged-files:put";
      assignmentKey: string;
      fileName: string;
      normalizedFileName: string;
      stagedAt: string;
      mimeType: string;
      canvasFileId?: string;
      blobBase64: string;
    }
  | {
      type: "staged-files:delete";
      assignmentKey: string;
      canvasFileId?: string;
      normalizedFileName?: string;
      stagedAt?: string;
    }
  | { type: "staged-files:list"; assignmentKey: string }
  | { type: "staged-files:clear-assignment"; assignmentKey: string }
  | {
      type: "staged-files:reconcile";
      assignmentKey: string;
      promotions: ReconcilePromotion[];
    }
  | {
      type: "staged-files:get-blob";
      assignmentKey: string;
      canvasFileId?: string;
      normalizedFileName?: string;
      stagedAt?: string;
    };

export type StagedFilesResponse =
  | { ok: true; files?: StagedFileRecord[]; blobBase64?: string; mimeType?: string }
  | { ok: false; error: string };
