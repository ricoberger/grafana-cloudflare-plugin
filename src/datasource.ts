import {
  DataFrame,
  DataSourceInstanceSettings,
  CoreApp,
  ScopedVars,
  DataQueryRequest,
  DataQueryResponse,
  LegacyMetricFindQueryOptions,
  MetricFindValue,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { lastValueFrom, Observable } from 'rxjs';

import { Query, Options, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<Query, Options> {
  constructor(instanceSettings: DataSourceInstanceSettings<Options>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<Query> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: Query, scopedVars: ScopedVars) {
    return {
      ...query,
      queryType: query.queryType || DEFAULT_QUERY.queryType,
      zone: getTemplateSrv().replace(query.zone, scopedVars),
    };
  }

  query(request: DataQueryRequest<Query>): Observable<DataQueryResponse> {
    return super.query(request);
  }

  async metricFindQuery(
    query: Query,
    options?: LegacyMetricFindQueryOptions,
  ): Promise<MetricFindValue[]> {
    const q = this.query({
      targets: [
        {
          ...query,
          refId: query.refId
            ? `metricsFindQuery-${query.refId}`
            : 'metricFindQuery',
        },
      ],
      range: options?.range,
    } as DataQueryRequest<Query>);

    const response = await lastValueFrom(q as Observable<DataQueryResponse>);

    if (
      response &&
      (!response.data.length || !response.data[0].fields.length)
    ) {
      return [];
    }

    return response
      ? (response.data[0] as DataFrame).fields[0].values.map((_, index) => {
        const name = (response.data[0] as DataFrame).fields[1].values[
          index
        ].toString();

        return {
          text: name,
          value: _.toString(),
        };
      })
      : [];
  }

  filterQuery(query: Query): boolean {
    if (query.queryType === 'metrics' && (!query.zone || !query.name)) {
      return false;
    }

    return true;
  }
}
