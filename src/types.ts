import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export const DEFAULT_QUERIES: Record<QueryType, Partial<Query>> = {
  zones: {},
  metrics: {
    name: 'httpRequests',
    aggregation: undefined,
    zone: '',
    filters: [{ field: '-', operator: '=', value: '' }],
    dimensions: [],
    orderBy: [],
    legend: '',
    limit: 100,
  },
};

export const DEFAULT_QUERY: Partial<Query> = {
  queryType: 'metrics',
  zone: '',
  limit: 100,
};

export type QueryType = 'zones' | 'metrics';

export interface Query extends DataQuery, QueryModelZones, QueryModelMetrics {
  queryType: QueryType;
}

interface QueryModelZones { }

interface QueryModelMetrics {
  name?: string;
  aggregation?: QueryModelMetricsAggregation;
  zone?: string;
  filters?: QueryModelMetricsFilter[];
  dimensions?: string[];
  orderBy?: string[];
  legend?: string;
  limit?: number;
}

export type QueryModelMetricsAggregation = 'sum' | 'avg' | 'count';

interface QueryModelMetricsFilter {
  field: string;
  operator: string;
  value: string;
}

export type OptionsAuthMethod = 'apiToken' | 'apiKey';

export interface Options extends DataSourceJsonData {
  authMethod?: OptionsAuthMethod;
  apiEmail?: string;
  zones?: Array<[string, string]>;
}

export interface OptionsSecure {
  apiToken?: string;
  apiKey?: string;
}
