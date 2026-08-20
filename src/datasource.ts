import {
  CoreApp,
  DataFrame,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  DataSourceWithSupplementaryQueriesSupport,
  LegacyMetricFindQueryOptions,
  MetricFindValue,
  QueryFixAction,
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

  // modifyQuery adds or removes a filter for the selected label / value when a
  // user clicks the filter buttons in the "Log details" section of a log line.
  // This is supported for the "workersLogs", "httpRequests" and
  // "firewallEvents" queries, which return logs and use the builder filters to
  // filter the log lines.
  modifyQuery(query: Query, action: QueryFixAction): Query {
    if (
      query.name !== 'workersLogs' &&
      query.name !== 'httpRequests' &&
      query.name !== 'firewallEvents'
    ) {
      return query;
    }

    const key = action.options?.key;
    const value = action.options?.value;
    if (!key || value === undefined) {
      return query;
    }

    let operator: string;
    switch (action.type) {
      case 'ADD_FILTER':
        operator = '=';
        break;
      case 'ADD_FILTER_OUT':
        operator = '!=';
        break;
      default:
        return query;
    }

    // Drop the placeholder filter (field "-"), which is used to display an
    // empty filter row in the UI without affecting the query, and remove any
    // existing filter for the same field, operator and value to avoid
    // duplicates.
    const filters = (query.filters ?? []).filter(
      (filter) =>
        filter.field !== '-' &&
        !(
          filter.field === key &&
          filter.operator === operator &&
          filter.value === value
        ),
    );

    filters.push({ field: key, operator, value });

    return { ...query, filterType: 'builder', filters };
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

    if (!response || !response.data.length) {
      return [];
    }

    // Filter value queries return one frame per value, where the value is
    // stored in the frame name (the frames do not contain any fields), so the
    // values are mapped from the frame names.
    if (query.queryType === 'filtervalues') {
      return response.data.map((data) => {
        return {
          text: data.name,
          value: data.name,
        };
      });
    }

    if (!response.data[0].fields.length) {
      return [];
    }

    return (response.data[0] as DataFrame).fields[0].values.map(
      (_, index) => {
        const name = (response.data[0] as DataFrame).fields[1].values[
          index
        ].toString();

        return {
          text: name,
          value: _.toString(),
        };
      },
    );
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
