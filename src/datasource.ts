import {
  DataFrame,
  DataSourceInstanceSettings,
  CoreApp,
  ScopedVars,
  DataQueryRequest,
  DataQueryResponse,
  LegacyMetricFindQueryOptions,
  MetricFindValue,
  SupplementaryQueryType,
  SupplementaryQueryOptions,
  DataSourceWithSupplementaryQueriesSupport,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { lastValueFrom, Observable } from 'rxjs';
import { cloneDeep } from 'lodash';

import { Query, Options, DEFAULT_QUERY } from './types';
import { VariableSupport } from './variablesupport';

export class DataSource
  extends DataSourceWithBackend<Query, Options>
  implements DataSourceWithSupplementaryQueriesSupport<Query> {
  constructor(instanceSettings: DataSourceInstanceSettings<Options>) {
    super(instanceSettings);
    this.variables = new VariableSupport(this);
  }

  getDefaultQuery(_: CoreApp): Partial<Query> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: Query, scopedVars: ScopedVars) {
    return {
      ...query,
      queryType: query.queryType || DEFAULT_QUERY.queryType,
      zone: getTemplateSrv().replace(query.zone, scopedVars),
      filter: getTemplateSrv().replace(query.filter, scopedVars),
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

  getSupportedSupplementaryQueryTypes(): SupplementaryQueryType[] {
    return [SupplementaryQueryType.LogsVolume];
  }

  getSupplementaryRequest(
    type: SupplementaryQueryType,
    request: DataQueryRequest<Query>,
    options?: SupplementaryQueryOptions,
  ): DataQueryRequest<Query> | undefined {
    if (!this.getSupportedSupplementaryQueryTypes().includes(type)) {
      return undefined;
    }

    const logsVolumeOption = { ...options, type };
    const logsVolumeRequest = cloneDeep(request);
    const targets = logsVolumeRequest.targets
      .map((query) => this.getSupplementaryQuery(logsVolumeOption, query))
      .filter((query): query is Query => query?.name === 'httpRequests');

    if (!targets.length) {
      return undefined;
    }

    return { ...logsVolumeRequest, targets };
  }

  getSupplementaryQuery(
    options: SupplementaryQueryOptions,
    query: Query,
  ): Query | undefined {
    if (query.hide) {
      return undefined;
    }

    switch (options.type) {
      case SupplementaryQueryType.LogsVolume: {
        return {
          ...query,
          queryType: 'logsvolume',
          refId: `log-volume-${query.refId}`,
        };
      }
      default:
        return undefined;
    }
  }
}
