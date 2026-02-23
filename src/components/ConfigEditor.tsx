import React, { ChangeEvent } from 'react';
import {
  FieldSet,
  IconButton,
  InlineField,
  InlineFieldRow,
  Input,
  RadioButtonGroup,
  SecretInput,
  useStyles2,
} from '@grafana/ui';
import {
  DataSourcePluginOptionsEditorProps,
  GrafanaTheme2,
} from '@grafana/data';
import { css } from '@emotion/css';

import { Options, OptionsAuthMethod, OptionsSecure } from '../types';

interface Props
  extends DataSourcePluginOptionsEditorProps<Options, OptionsSecure> { }

export function ConfigEditor(props: Props) {
  const styles = useStyles2((theme: GrafanaTheme2) => ({
    marginTop: css`
      margin-top: ${theme.spacing(4)};
    `,
  }));

  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  return (
    <>
      <InlineField
        data-testid="authentication-method"
        label="Authentication Method"
        labelWidth={25}
      >
        <RadioButtonGroup<OptionsAuthMethod>
          options={[
            { label: 'API Token', value: 'apiToken' },
            { label: 'API Key', value: 'apiKey' },
          ]}
          value={options.jsonData.authMethod}
          onChange={(value: OptionsAuthMethod) => {
            onOptionsChange({
              ...options,
              jsonData: {
                ...options.jsonData,
                authMethod: value,
              },
            });
          }}
        />
      </InlineField>

      {jsonData.authMethod === 'apiToken' && (
        <>
          <InlineField label="API Token" labelWidth={25} interactive>
            <SecretInput
              required
              isConfigured={secureJsonFields.apiToken}
              value={secureJsonData?.apiToken}
              width={40}
              onReset={() => {
                onOptionsChange({
                  ...options,
                  secureJsonFields: {
                    ...options.secureJsonFields,
                    apiToken: false,
                  },
                  secureJsonData: {
                    ...options.secureJsonData,
                    apiToken: '',
                  },
                });
              }}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                onOptionsChange({
                  ...options,
                  secureJsonData: {
                    apiToken: event.target.value,
                  },
                });
              }}
            />
          </InlineField>
        </>
      )}

      {jsonData.authMethod === 'apiKey' && (
        <>
          <InlineField label="API Email" labelWidth={25} interactive>
            <Input
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                onOptionsChange({
                  ...options,
                  jsonData: {
                    ...jsonData,
                    apiEmail: event.target.value,
                  },
                });
              }}
              value={jsonData.apiEmail}
              width={40}
            />
          </InlineField>
          <InlineField label="API Key" labelWidth={25} interactive>
            <SecretInput
              required
              isConfigured={secureJsonFields.apiKey}
              value={secureJsonData?.apiKey}
              width={40}
              onReset={() => {
                onOptionsChange({
                  ...options,
                  secureJsonFields: {
                    ...options.secureJsonFields,
                    apiKey: false,
                  },
                  secureJsonData: {
                    ...options.secureJsonData,
                    apiKey: '',
                  },
                });
              }}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                onOptionsChange({
                  ...options,
                  secureJsonData: {
                    apiKey: event.target.value,
                  },
                });
              }}
            />
          </InlineField>
        </>
      )}

      <FieldSet className={styles.marginTop} label="Zones">
        <IconButton
          name="plus"
          aria-label="Add Zone"
          onClick={() => {
            onOptionsChange({
              ...options,
              jsonData: {
                ...jsonData,
                zones: jsonData.zones
                  ? [...jsonData.zones, ['', '']]
                  : [['', '']],
              },
            });
          }}
        />

        {jsonData?.zones?.map((zone, index) => (
          <InlineFieldRow key={index}>
            <InlineField label="ID" labelWidth={10}>
              <Input
                width={40}
                value={zone[0]}
                onChange={(e) => {
                  const newZones = [...jsonData.zones!];
                  newZones[index] = [
                    e.currentTarget.value,
                    jsonData.zones![index][1],
                  ];

                  onOptionsChange({
                    ...options,
                    jsonData: {
                      ...jsonData,
                      zones: newZones,
                    },
                  });
                }}
              />
            </InlineField>
            <InlineField label="Name" labelWidth={10}>
              <Input
                width={40}
                value={zone[1]}
                onChange={(e) => {
                  const newZones = [...jsonData.zones!];
                  newZones[index] = [
                    jsonData.zones![index][0],
                    e.currentTarget.value,
                  ];

                  onOptionsChange({
                    ...options,
                    jsonData: {
                      ...jsonData,
                      zones: newZones,
                    },
                  });
                }}
              />
            </InlineField>
            <IconButton
              name="trash-alt"
              aria-label="Remove Zone"
              onClick={(e) => {
                e.preventDefault();
                const newZones = [...jsonData.zones!];
                newZones.splice(index, 1);

                onOptionsChange({
                  ...options,
                  jsonData: {
                    ...jsonData,
                    zones: newZones,
                  },
                });
              }}
            />
          </InlineFieldRow>
        ))}
      </FieldSet>
    </>
  );
}
