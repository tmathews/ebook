import argparse
import uuid
import shutil
import os
import sys
import pybars
import datetime
import dateutil.parser
import tempfile

def filter_images (x):
	_, ext = os.path.splitext(x)
	return ext in ['.jpg', '.png']

def file_to_string (path):
	file = open(path, 'r')
	buf = file.read()
	file.close()
	return buf

def string_to_file (string, path):
	f = open(path, 'w')
	f.write(string)
	f.close()

def replace_template (compiler, path, scope):
	a_str = file_to_string(path)
	tmpl = compiler.compile(a_str)
	buf = tmpl(scope)
	string_to_file(buf, path)

def get_date (x):
	date = datetime.datetime.now()
	if x:
		date = dateutil.parser.parse(x)
	xstr = date.isoformat(timespec='seconds').replace('+00:00', 'Z')
	if xstr[len(xstr) - 1] != 'Z':
		xstr += 'Z'
	return xstr

def get_writing_mode (ltr):
	if ltr:
		return "horizontal-lr"
	else:
		return "horizontal-rl"

def parse_comma_list (xs):
	if isinstance(xs, str):
		return list(map(lambda x: x.strip(), xs.split(",")))
	return xs

def create_epub (args):
	src_dir = args.dir + os.sep

	if not os.path.exists(src_dir):
		print("Specified path does not exist")
		sys.exit(1)
		return

	if not os.path.isdir(src_dir):
		print("Specified path is not a directory.")
		sys.exit(2)
		return

	compiler = pybars.Compiler()

	book_id = args.uuid or uuid.uuid1()
	title = args.title or os.path.basename(os.path.dirname(src_dir))
	numbers = []

	working_dir = os.path.join(tempfile.gettempdir(), str(book_id))
	oebps_dir = os.path.join(working_dir, 'OEBPS')
	img_dir = os.path.join(oebps_dir, 'Images')
	txt_dir = os.path.join(oebps_dir, 'Text')

	# Create working directory from template
	shutil.copytree(os.path.join('.', 'template'), working_dir)

	# Copy over image files from the source directory
	files = list(filter(filter_images, os.listdir(src_dir)))
	counter = 0
	fill = len(str(len(files)))
	xhtml_path = os.path.join(txt_dir, 'template.xhtml')
	xhtml = file_to_string(xhtml_path)
	xhtml_template = compiler.compile(xhtml)
	os.remove(xhtml_path)
	for f in files:
		_, ext = os.path.splitext(f)
		counter += 1
		fill_name = str(counter).zfill(fill)
		file_name = fill_name + ext
		dic = {
			"number": fill_name,
			"file_name": file_name
		}
		shutil.copy(os.path.join(src_dir, f), os.path.join(img_dir, file_name))
		numbers.append(dic)

		# Create a text page for the image
		buf = xhtml_template(dic)
		string_to_file(buf, os.path.join(txt_dir, fill_name + '.xhtml'))

		# Create a cover image from the first image.
		if counter == 1:
			shutil.copy(os.path.join(src_dir, f),
				os.path.join(img_dir, 'cover' + ext))

	# Write content.opf
	replace_template(compiler, os.path.join(oebps_dir, 'content.opf'), {
		"book_id": book_id,
		"title": title,
		"contributor": args.contributor,
		"creator": args.creator,
		"language": args.language,
		"date_modified": get_date(args.date_modified),
		"writing_mode": get_writing_mode(args.ltr),
		"direction": 'ltr' if args.ltr else 'rtl',
		"numbers": numbers,
	})

	# Write nav.xhtml
	replace_template(compiler, os.path.join(oebps_dir, 'nav.xhtml'), {
		"title": title,
		"first_page": numbers[0]['number']
	})

	# Write toc.ncx
	replace_template(compiler, os.path.join(oebps_dir, 'toc.ncx'), {
		"book_id": book_id,
		"title": title,
		"first_page": numbers[0]['number']
	})

	# Zip it as epub to the output location
	output_dir = os.getcwd()
	if args.output: output_dir = os.path.dirname(args.output)
	output_name = os.path.basename(args.output or os.path.dirname(src_dir))
	if not output_name:
		output_name = os.path.basename(os.path.dirname(src_dir))
	output_name, output_ext = os.path.splitext(output_name)
	if not output_ext:
		output_ext = ".epub"
	output_path = os.path.join(output_dir, output_name)
	shutil.make_archive(output_path, 'zip', working_dir)
	shutil.move(output_path + '.zip', output_path + output_ext)
	shutil.rmtree(working_dir)

def default_value_parser (x):
	return x

def input_kv (dic, key, value_parser=default_value_parser):
	dic[key] = input("    %s (%s): " % (key, dic[key])) or dic[key]

def get_input_metadata (dic):
	print("Please provide ebook details.\n")
	input_kv(dic, "uuid")
	input_kv(dic, "title")
	input_kv(dic, "date_modified")
	input_kv(dic, "description")
	input_kv(dic, "language")
	input_kv(dic, "contributor")
	input_kv(dic, "creator")
	input_kv(dic, "subjects", parse_comma_list)

if __name__ == "__main__":
	parser = argparse.ArgumentParser(description="Creates an epub from a directory.")
	parser.add_argument("dir",
		help="The source directory to read from.")
	parser.add_argument("--output",
		help="The desired output file path.")
	parser.add_argument("--interactive",
		action="store_true",
		help="Require input for metadata.")
	parser.add_argument("--rtl",
		help="Right to left reading mode.",
		action="store_false",
		dest="ltr")
	parser.add_argument("--uuid", default="")
	parser.add_argument("--title", default="")
	parser.add_argument("--description", default="")
	parser.add_argument("--language", default="en")
	parser.add_argument("--contributor", default="")
	parser.add_argument("--creator", default="")
	parser.add_argument("--date-modified",
		help="Current date will be used if not provided.",
		default="")
	parser.add_argument("--subjects",
		nargs="*",
		help="A comma separated list.",
		default=[])
	args = parser.parse_args()
	if args.interactive:
		get_input_metadata(vars(args))
		print("""
Review new details.

    uuid: %s
    title: %s
    modified: %s
    description: %s
    language: %s
    contributor: %s
    creator: %s
    subjects: %s

Type "no" to abort.""" % (
		args.uuid or "(autogen)",
		args.title or "(use directory name)",
		args.date_modified or "(now)",
		args.description,
		args.language,
		args.contributor,
		args.creator,
		args.subjects))
		if input() == "no":
			sys.exit(3)
	create_epub(args)