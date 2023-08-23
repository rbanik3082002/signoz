import { ChartData } from 'chart.js';
import { GraphOnClickHandler, ToggleGraphProps } from 'components/Graph/types';
import { MutableRefObject, ReactNode } from 'react';
import { Layout } from 'react-grid-layout';
import { UseQueryResult } from 'react-query';
import { ErrorResponse, SuccessResponse } from 'types/api';
import { Widgets } from 'types/api/dashboard/getAll';
import { MetricRangePayloadProps } from 'types/api/metrics/getQueryRange';

import { MenuItemKeys } from '../WidgetHeader/contants';
import { LegendEntryProps } from './FullView/types';

export interface GraphVisibilityLegendEntryProps {
	graphVisibilityStates: boolean[];
	legendEntry: LegendEntryProps[];
}

export interface WidgetGraphComponentProps {
	enableModel: boolean;
	enableWidgetHeader: boolean;
	widget: Widgets;
	queryResponse: UseQueryResult<
		SuccessResponse<MetricRangePayloadProps> | ErrorResponse
	>;
	errorMessage: string | undefined;
	data: ChartData;
	name: string;
	setLayout?: (layout: Layout[]) => void;
	onDragSelect?: (start: number, end: number) => void;
	onClickHandler?: GraphOnClickHandler;
	threshold?: ReactNode;
	headerMenuList: MenuItemKeys[];
}

export interface GridCardGraphProps {
	widget: Widgets;
	name: string;
	setLayout?: WidgetGraphComponentProps['setLayout'];
	onDragSelect?: (start: number, end: number) => void;
	onClickHandler?: GraphOnClickHandler;
	threshold?: ReactNode;
	headerMenuList?: WidgetGraphComponentProps['headerMenuList'];
	isQueryEnabled: boolean;
}

export interface GetGraphVisibilityStateOnLegendClickProps {
	data: ChartData;
	isExpandedName: boolean;
	name: string;
}

export interface ToggleGraphsVisibilityInChartProps {
	graphsVisibilityStates: GraphVisibilityLegendEntryProps['graphVisibilityStates'];
	lineChartRef: MutableRefObject<ToggleGraphProps | undefined>;
}