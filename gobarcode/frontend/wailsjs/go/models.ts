export namespace excel {
	
	export class LabelInfo {
	    sheet_map: Record<number, string>;
	    selected_sheet: number;
	    selected_sheet_name: string;
	    header_row: number;
	    header_col: string;
	    header_row_values: string[];
	    error: any;
	
	    static createFrom(source: any = {}) {
	        return new LabelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sheet_map = source["sheet_map"];
	        this.selected_sheet = source["selected_sheet"];
	        this.selected_sheet_name = source["selected_sheet_name"];
	        this.header_row = source["header_row"];
	        this.header_col = source["header_col"];
	        this.header_row_values = source["header_row_values"];
	        this.error = source["error"];
	    }
	}

}

