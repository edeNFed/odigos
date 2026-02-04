export interface AppID {
  namespace: string;
  kind: string;
  name: string;
}

export interface AppInfo {
  appId: AppID;
  bucketCount: number;
  totalSamples: number;
  oldestData: string;
  newestData: string;
}

export interface BucketInfo {
  startTime: string;
  endTime: string;
  sampleCount: number;
}

export interface FlameGraphNode {
  name: string;
  value: number;
  children?: FlameGraphNode[];
}

export interface AppsResponse {
  apps: AppInfo[];
}

export interface BucketsResponse {
  buckets: BucketInfo[];
}
