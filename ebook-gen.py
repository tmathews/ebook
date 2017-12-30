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
		"title": title,
		"book_id": book_id,
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

if __name__ == "__main__":
	parser = argparse.ArgumentParser(description="Creates an epub from a directory.")
	parser.add_argument("dir",
		help="The source directory to read from.")
	parser.add_argument("--output",
		help="The desired output file path.")
	parser.add_argument("--rtl",
		help="Right to left reading mode.",
		action="store_false",
		dest="ltr")
	parser.add_argument("--uuid")
	parser.add_argument("--title")
	parser.add_argument("--language", default="en-US")
	parser.add_argument("--contributor", default="Unknown Contributor")
	parser.add_argument("--creator", default="Unknown Author")
	parser.add_argument("--date-modified")
	args = parser.parse_args()
	create_epub(args)