class Overloading{
	public int add(int a, int b){
		return a+b;
	}

	public int add(int a, int b, int c){
		return a+b+c;
	}

	public static void main(String[] args){
		Overloading ov =  new Overloading();

		int result1 = ov.add(12, 3);
		int result2 = ov.add(12, 3, 4);

		System.out.println(result1);
		System.out.println(result2);
	}
}
