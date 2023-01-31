export interface IComment {
  ID: string;
  content: string;
  author_id: string;
  created_at: string;
  updated_at: string;
  parent_id: string;
  vote_pos_count: number; // Excludes users own vote
  vote_neg_count: number; // Excludes users own vote
  my_vote: null | {
    uid: string;
    is_upvote: boolean;
  };
}

export interface IPostCard {
  ID: string;
  author_id: string;
  title: string;
  description: string;
  tags: string[];
  created_at: string;
  updated_at: string;
  slug: string;
  img_blur: string;
  vote_pos_count: number; // Excludes users own vote
  vote_neg_count: number; // Excludes users own vote
  my_vote: null | {
    uid: string;
    is_upvote: boolean;
  };
  img_url: string; //img_url is stored here so that rerender can be triggered when the image is updated by modifying the query string
}

export interface IPost extends IPostCard {
  body: string;
}
