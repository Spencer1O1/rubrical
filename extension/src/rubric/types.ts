export type RubricRating = {
  title: string;
  description: string;
  points: string;
};

export type RubricTableRow = {
  criterion: string;
  criterionLongDescription: string;
  ratings: RubricRating[];
  points: string;
};

export type RubricTable = {
  header: string[];
  rows: RubricTableRow[];
};

export const DEFAULT_HEADER = ["Criteria", "Ratings", "Points"];
