import {
  CoreApp,
  DataFrame,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  DataSourceWithSupplementaryQueriesSupport,
  LegacyMetricFindQueryOptions,
  MetricFindValue,
  ScopedVars,
  SupplementaryQueryOptions,
  SupplementaryQueryType,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { cloneDeep } from 'lodash';
import { lastValueFrom, Observable } from 'rxjs';

import { DEFAULT_QUERY, Options, Query } from './types';
import { isAccountScopedMetric } from './utils';
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
    const filters = query.filters?.map((filter) => ({
      field: filter.field,
      operator: filter.operator,
      value: getTemplateSrv().replace(filter.value, scopedVars),
    }));

    return {
      ...query,
      queryType: query.queryType || DEFAULT_QUERY.queryType,
      zone: getTemplateSrv().replace(query.zone, scopedVars),
      filter: getTemplateSrv().replace(query.filter, scopedVars),
      filters: filters,
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

    if (query.queryType === 'filtervalues') {
      return response
        ? response.data.map((data) => {
          return {
            text: data.name,
            value: data.name,
          };
        })
        : [];
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
    if (
      query.queryType === 'filtervalues' &&
      (!query.name ||
        !query.field ||
        (!isAccountScopedMetric(query.name) && !query.zone))
    ) {
      return false;
    }

    if (
      query.queryType === 'metrics' &&
      (!query.name || (!isAccountScopedMetric(query.name) && !query.zone))
    ) {
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
      .filter(
        (query): query is Query =>
          query?.name === 'httpRequests' ||
          query?.name === 'firewallEvents' ||
          query?.name === 'workersLogs',
      );

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
